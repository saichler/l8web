package server

import (
	"fmt"
	"io"
	"net/http"
	"sync"
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
	resources ifs.IResources
	adjacents []ifs.IResources
}

var mtx = &sync.Mutex{}
var registered = map[uint32]bool{}
var registeredAuth = false
var authEnabled = false
var adjacentTokens = make(map[string]string)

func (this *WebService) Activate(sla *ifs.ServiceLevelAgreement, vnic ifs.IVNic) error {
	this.resources = vnic.Resources()
	vnic.Resources().Registry().Register(&l8web.L8WebService{})
	this.server = sla.Args()[0].(ifs.IWebServer)
	go func() {
		time.Sleep(time.Second * 2)
		fmt.Println("Sending Get Multicast for EndPoints ", vnic.Resources().SysConfig().VnetPort)
		vnic.Multicast(health.ServiceName, 0, ifs.EndPoints, nil)
	}()

	mtx.Lock()
	defer mtx.Unlock()

	if !registeredAuth {
		registeredAuth = true
		http.DefaultServeMux.HandleFunc("/auth", this.Auth)
		http.DefaultServeMux.HandleFunc("/registry", this.Registry)
		http.DefaultServeMux.HandleFunc("/tfaSetup", this.TFASetup)
		http.DefaultServeMux.HandleFunc("/tfaSetupVerify", this.TFAVerify)
		http.DefaultServeMux.HandleFunc("/tfaVerify", this.TFAVerify)
	}

	for _, n := range sla.Args() {
		nic, ok := n.(ifs.IVNic)
		if ok {
			_, ok = registered[nic.Resources().SysConfig().VnetPort]
			if !ok {
				if this.adjacents == nil {
					this.adjacents = make([]ifs.IResources, 0)
				}
				this.adjacents = append(this.adjacents, nic.Resources())
				registered[nic.Resources().SysConfig().VnetPort] = true
				go func() {
					time.Sleep(time.Second * 5)
					nic.Resources().Services().Activate(sla, nic)
				}()
			}
		}
	}
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
	token, err := this.resources.Security().Authenticate(user.User, user.Pass)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		authToken := &l8api.AuthToken{}
		authToken.Error = err.Error()
		jsn, _ := protojson.Marshal(authToken)
		w.Write(jsn)
		fmt.Println("Failed to authenticate user/pass #3")
		return
	}

	//We need to authenticate with the adjacent as well
	//This is a temp solution, need to integrate it.
	if this.adjacents != nil {
		for _, adjacent := range this.adjacents {
			aToken, aErr := adjacent.Security().Authenticate(user.User, user.Pass)
			if aErr == nil {
				mtx.Lock()
				adjacentTokens[token] = aToken
				mtx.Unlock()
			}
		}
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

func (this *WebService) Registry(w http.ResponseWriter, r *http.Request) {
	if authEnabled {
		bearer := r.Header.Get("Authorization")
		if bearer == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, ok := this.resources.Security().ValidateToken(bearer)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}
	typeList := this.resources.Registry().TypeList()
	byt, _ := protojson.Marshal(typeList)
	w.WriteHeader(http.StatusOK)
	w.Write(byt)
}
