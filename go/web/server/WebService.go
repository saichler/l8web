package server

import (
	"time"

	"github.com/saichler/l8srlz/go/serialize/object"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types/l8web"
	"github.com/saichler/l8utils/go/utils/web"
	"github.com/saichler/l8bus/go/overlay/health"
	"github.com/saichler/l8bus/go/overlay/plugins"
)

const (
	ServiceTypeName = "WebService"
)

type WebService struct {
	server ifs.IWebServer
}

func (this *WebService) Activate(serviceName string, serviceArea byte,
	resources ifs.IResources, listener ifs.IServiceCacheListener, args ...interface{}) error {
	resources.Registry().Register(&l8web.L8WebService{})
	this.server = args[0].(ifs.IWebServer)
	vnic, ok := listener.(ifs.IVNic)
	if ok {
		go func() {
			time.Sleep(time.Second * 2)
			vnic.Resources().Logger().Info("Sending Get Multicast for EndPoints")
			vnic.Multicast(health.ServiceName, 0, ifs.EndPoints, nil)
		}()
	}
	return nil
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
