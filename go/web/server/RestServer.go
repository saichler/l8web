package server

import (
	"bytes"
	"fmt"
	"github.com/saichler/layer8/go/overlay/protocol"
	"github.com/saichler/shared/go/share/certs"
	"net/http"
	"os"
	"strconv"
	"time"
)

type RestServer struct {
	webServer *http.Server
	RestServerConfig
}

type RestServerConfig struct {
	Host           string
	Port           int
	CertName       string
	Authentication bool
	Prefix         string
}

func NewRestServer(config *RestServerConfig) (*RestServer, error) {
	rs := &RestServer{}
	rs.Authentication = config.Authentication
	rs.CertName = config.CertName
	rs.Host = config.Host
	rs.Port = config.Port
	rs.Prefix = config.Prefix

	if rs.CertName != "" {
		_, err := os.Open(rs.CertName + ".crt")
		if err != nil {
			return rs, certs.CreateLayer8Crt(rs.CertName, protocol.MachineIP, int64(rs.Port))
		}
	}
	return rs, nil
}

func (this *RestServer) patternOf(handler *ServicePointHandler) string {
	buff := bytes.Buffer{}
	buff.WriteString(this.Prefix)
	buff.WriteString(strconv.Itoa(int(handler.serviceArea)))
	buff.WriteString("/")
	buff.WriteString(handler.serviceName)
	fmt.Println("Server Path=", buff.String())
	return buff.String()
}

func (this *RestServer) AddServicePointHandler(handler *ServicePointHandler) {
	http.DefaultServeMux.HandleFunc(this.patternOf(handler), handler.serveHttp)
}

func (this *RestServer) Start() error {
	var err error
	this.webServer = &http.Server{
		Addr:    this.Host + ":" + strconv.Itoa(this.Port),
		Handler: http.DefaultServeMux,
	}
	if this.CertName != "" {
		err = this.webServer.ListenAndServeTLS(this.CertName+".crt", this.CertName+".crtKey")
	} else {
		err = this.webServer.ListenAndServe()
	}
	return err
}

func (this *RestServer) Stop() {
	this.webServer.Shutdown(this)
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
