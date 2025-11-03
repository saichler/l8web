package server

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/saichler/l8bus/go/overlay/health"
	"github.com/saichler/l8bus/go/overlay/plugins"
	"github.com/saichler/l8srlz/go/serialize/object"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types/l8api"
	"github.com/saichler/l8types/go/types/l8web"
	"github.com/saichler/l8utils/go/utils/web"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	ServiceTypeName = "WebService"
)

type WebService struct {
	server    ifs.IWebServer
	resources []ifs.IResources
}

func (this *WebService) Activate(sla *ifs.ServiceLevelAgreement, vnic1 ifs.IVNic) error {
	this.resources = make([]ifs.IResources, 0)
	this.resources = append(this.resources, vnic1.Resources())
	vnics := make([]ifs.IVNic, 0)
	vnics = append(vnics, vnic1)
	if len(sla.Args()) > 1 {
		for i := 1; i < len(sla.Args()); i++ {
			nic, ok := sla.Args()[i].(ifs.IVNic)
			if ok {
				vnics = append(vnics, nic)
				this.resources = append(this.resources, nic.Resources())
				nic.Resources().Registry().Register(&l8web.L8WebService{})
			}
		}
	}
	vnic1.Resources().Registry().Register(&l8web.L8WebService{})

	this.server = sla.Args()[0].(ifs.IWebServer)
	go func() {
		time.Sleep(time.Second * 2)
		for _, vnic := range vnics {
			vnic.Resources().Logger().Info("Sending Get Multicast for EndPoints")
			vnic.Multicast(health.ServiceName, 0, ifs.EndPoints, nil)
		}
	}()
	http.DefaultServeMux.HandleFunc("/auth", this.Auth)
	return nil
}

func (this *WebService) Auth(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to read user/pass #1"))
		w.Write([]byte(err.Error()))
		fmt.Println("Failed to read user/pass #1")
		return
	}
	user := &l8api.AuthUser{}
	err = protojson.Unmarshal(data, user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to read user/pass #2"))
		w.Write([]byte(err.Error()))
		fmt.Println("Failed to read user/pass #2")
		return
	}
	token, err := this.resources[0].Security().Authenticate(user.User, user.Pass)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		authToken := &l8api.AuthToken{}
		authToken.Error = err.Error()
		jsn, _ := protojson.Marshal(authToken)
		w.Write(jsn)
		fmt.Println("Failed to authenticate user/pass #3")
		return
	}

	authToken := &l8api.AuthToken{}
	authToken.Token = token
	w.WriteHeader(http.StatusOK)
	jsn, _ := protojson.Marshal(authToken)
	w.Write(jsn)
}

func (this *WebService) DeActivate() error {
	return nil
}

func (this *WebService) Post(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	webService := pb.Element().(*l8web.L8WebService)
	ws := &web.WebService{}
	ws.DeSerialize(webService)
	vnic.Resources().Logger().Info("Received Webservice ", ws.ServiceName(), " ", ws.ServiceArea())
	if ws.Plugin() != "" {
		plg := &l8web.L8Plugin{Data: ws.Plugin()}
		plugins.LoadPlugin(plg, vnic)
	}
	this.server.RegisterWebService(ws, vnic)
	return object.New(nil, nil)
}

func (this *WebService) Put(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	return nil
}
func (this *WebService) Patch(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	return nil
}
func (this *WebService) Delete(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	return nil
}
func (this *WebService) GetCopy(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	return nil
}
func (this *WebService) Get(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	return object.New(nil, nil)
}
func (this *WebService) Failed(pb ifs.IElements, vnic ifs.IVNic, msg *ifs.Message) ifs.IElements {
	return nil
}
func (this *WebService) TransactionConfig() ifs.ITransactionConfig {
	return nil
}
func (this *WebService) WebService() ifs.IWebService {
	return nil
}
