package server

import (
	"github.com/saichler/l8srlz/go/serialize/object"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types"
	"github.com/saichler/l8utils/go/utils/web"
	"github.com/saichler/layer8/go/overlay/health"
	"github.com/saichler/layer8/go/overlay/plugins"
	"time"
)

const (
	ServiceName     = "WebEndPoints"
	ServiceTypeName = "WebEndPointsService"
)

type WebEndPointsService struct {
	server ifs.IWebServer
}

func (this *WebEndPointsService) Activate(serviceName string, serviceArea uint16,
	resources ifs.IResources, listener ifs.IServiceCacheListener, args ...interface{}) error {
	resources.Registry().Register(&types.WebService{})
	this.server = args[0].(ifs.IWebServer)
	vnic, ok := listener.(ifs.IVNic)
	if ok {
		go func() {
			time.Sleep(time.Second * 10)
			vnic.Resources().Logger().Info("Sending Get Multicast for EndPoints")
			vnic.Multicast(health.ServiceName, 0, ifs.EndPoints, nil)
		}()
	}
	return nil
}

func (this *WebEndPointsService) DeActivate() error {
	return nil
}

func (this *WebEndPointsService) Post(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	webService := pb.Element().(*types.WebService)
	ws := &web.WebService{}
	ws.DeSerialize(webService)
	vnic.Resources().Logger().Info("Received Webservice ", ws.ServiceName(), " ", ws.ServiceArea())
	if ws.Plugin() != "" {
		plg := &types.Plugin{Data: ws.Plugin()}
		plugins.LoadPlugin(plg, vnic)
	}
	this.server.RegisterWebService(ws, vnic)
	return object.New(nil, nil)
}

func (this *WebEndPointsService) Put(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	return nil
}
func (this *WebEndPointsService) Patch(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	return nil
}
func (this *WebEndPointsService) Delete(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	return nil
}
func (this *WebEndPointsService) GetCopy(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	return nil
}
func (this *WebEndPointsService) Get(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	return object.New(nil, nil)
}
func (this *WebEndPointsService) Failed(pb ifs.IElements, vnic ifs.IVNic, msg ifs.IMessage) ifs.IElements {
	return nil
}
func (this *WebEndPointsService) TransactionMethod() ifs.ITransactionMethod {
	return nil
}
func (this *WebEndPointsService) WebService() ifs.IWebService {
	return nil
}
