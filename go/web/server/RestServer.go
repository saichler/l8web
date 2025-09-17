package server

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8utils/go/utils/certs"
	"github.com/saichler/l8utils/go/utils/maps"
	"github.com/saichler/layer8/go/overlay/protocol"
	"google.golang.org/protobuf/proto"
)

var endPoints = maps.NewSyncMap()

type RestServer struct {
	webServer    *http.Server
	webServerDev *http.Server
	RestServerConfig
}

type RestServerConfig struct {
	Host           string
	Port           int
	CertName       string
	Authentication bool
	Prefix         string
}

func NewRestServer(config *RestServerConfig) (ifs.IWebServer, error) {
	rs := &RestServer{}
	rs.Authentication = config.Authentication
	rs.CertName = config.CertName
	rs.Host = config.Host
	rs.Port = config.Port
	rs.Prefix = config.Prefix

	http.DefaultServeMux = http.NewServeMux()
	rs.LoadWebUI()

	if rs.CertName != "" {
		//For development
		_, err := os.Open(rs.CertName + ".crt")
		if err != nil {
			fmt.Println("Error loading dev certificate:", err)
			certs.CreateLayer8Crt(rs.CertName+"-dev", protocol.MachineIP, int64(rs.Port+2000))
		}
		_, err = os.Open(rs.CertName + ".crt")
		if err != nil {
			fmt.Println("Error loading certificate:", err)
			certs.CreateLayer8Crt(rs.CertName, protocol.MachineIP, int64(rs.Port))
		}
	}

	return rs, nil
}

func (this *RestServer) patternOf(handler *ServiceHandler) string {
	buff := bytes.Buffer{}
	buff.WriteString(this.Prefix)
	buff.WriteString(strconv.Itoa(int(handler.serviceArea)))
	buff.WriteString("/")
	buff.WriteString(handler.serviceName)
	fmt.Println("Server Path=", buff.String())
	return buff.String()
}

func (this *RestServer) RegisterWebService(ws ifs.IWebService, vnic ifs.IVNic) {
	handler := &ServiceHandler{}
	handler.serviceName = ws.ServiceName()
	handler.serviceArea = ws.ServiceArea()
	handler.vnic = vnic
	handler.method2Body = make(map[string]proto.Message)
	handler.method2Resp = make(map[string]proto.Message)

	handler.addEndPoint(http.MethodPost, ws.PostBody(), ws.PostResp())
	handler.addEndPoint(http.MethodPut, ws.PutBody(), ws.PutResp())
	handler.addEndPoint(http.MethodPatch, ws.PatchBody(), ws.PatchResp())
	handler.addEndPoint(http.MethodDelete, ws.DeleteBody(), ws.DeleteResp())
	handler.addEndPoint(http.MethodGet, ws.GetBody(), ws.GetResp())

	path := this.patternOf(handler)
	_, ok := endPoints.Get(path)
	if !ok {
		endPoints.Put(path, true)
		fmt.Println("Registering path=", path)
		http.DefaultServeMux.HandleFunc(this.patternOf(handler), handler.serveHttp)
	}
}

func (this *RestServer) Start() error {
	var err error
	this.webServer = &http.Server{
		Addr:    this.Host + ":" + strconv.Itoa(this.Port),
		Handler: http.DefaultServeMux,
	}
	this.webServerDev = &http.Server{
		Addr:    this.Host + ":" + strconv.Itoa(this.Port+2000),
		Handler: http.DefaultServeMux,
	}

	if this.CertName != "" {
		//For development
		go func() {
			err = this.webServer.ListenAndServeTLS(this.CertName+"-dev.crt", this.CertName+"-dev.crtKey")
			if err != nil {
				fmt.Println("Error starting dev web server ", err)
			}
		}()
		err = this.webServer.ListenAndServeTLS(this.CertName+".crt", this.CertName+".crtKey")
		if err != nil && !strings.Contains(err.Error(), "Server closed") {
			fmt.Println("Error starting web server ", err)
			this.CertName = ""
			err = this.webServer.ListenAndServe()
		}
	} else {
		err = this.webServer.ListenAndServe()
	}
	return err
}

func (this *RestServer) Stop() {
	this.webServer.Shutdown(this)
	endPoints.Clean()
	fmt.Println("Cleaned!")
}

func (this *RestServer) Deadline() (deadline time.Time, ok bool) {
	return time.Now(), true
}

func (this *RestServer) Done() <-chan struct{} {
	return nil
}

func (this *RestServer) Err() error {
	return nil
}

func (this *RestServer) Value(key interface{}) interface{} {
	return nil
}
