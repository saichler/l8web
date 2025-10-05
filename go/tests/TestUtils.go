package tests

import (
	"encoding/base64"
	"os"
	"testing"
	"time"

	"github.com/saichler/l8bus/go/overlay/plugins"
	"github.com/saichler/l8bus/go/overlay/protocol"
	"github.com/saichler/l8bus/go/overlay/vnic"
	. "github.com/saichler/l8test/go/infra/t_resources"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types/l8api"
	"github.com/saichler/l8types/go/types/l8health"
	"github.com/saichler/l8types/go/types/l8web"
	"github.com/saichler/l8web/go/web/client"
	"github.com/saichler/l8web/go/web/server"
)

const (
	VNET_PORT = 28000
)

func createWebServer(t *testing.T) (ifs.IVNic, ifs.IWebServer, bool) {
	resources, _ := CreateResources(VNET_PORT, 1, ifs.Info_Level)
	webNic := vnic.NewVirtualNetworkInterface(resources, nil)
	webNic.Start()
	webNic.WaitForConnection()

	webNic.Resources().Registry().Register(&l8web.L8Empty{})
	webNic.Resources().Registry().Register(&l8health.L8Top{})

	serverConfig := &server.RestServerConfig{
		Host:           protocol.MachineIP,
		Port:           8080,
		Authentication: true,
		CertName:       "test",
		Prefix:         "/test/",
	}
	srv, err := server.NewRestServer(serverConfig)
	if err != nil {
		Log.Fail(t, err)
		return nil, srv, false
	}
	webNic.Resources().Services().RegisterServiceHandlerType(&server.WebService{})
	_, err = webNic.Resources().Services().Activate(server.ServiceTypeName, ifs.WebService,
		0, webNic.Resources(), webNic, srv)
	if err != nil {
		Log.Fail(t, err.Error())
		return nil, srv, false
	}
	go srv.Start()
	return webNic, srv, true
}

func createServiceNic(t *testing.T) (ifs.IVNic, bool) {
	resources, _ := CreateResources(VNET_PORT, 2, ifs.Info_Level)
	serviceNic := vnic.NewVirtualNetworkInterface(resources, nil)
	serviceNic.Start()
	serviceNic.WaitForConnection()

	err := PushPlugin(serviceNic, "service.so")
	if err != nil {
		Log.Fail(t, err.Error())
		return nil, false
	}
	time.Sleep(time.Second * 2)
	return serviceNic, true
}

func createRestClient(t *testing.T, pb interface{}) (*client.RestClient, bool) {
	return createRestClient2(t, pb, "/test/")
}

func createRestClient2(t *testing.T, pb interface{}, prefix string) (*client.RestClient, bool) {
	resources, _ := CreateResources(VNET_PORT, 3, ifs.Info_Level)
	resources.Registry().Register(&l8web.L8Empty{})
	resources.Registry().Register(&l8api.AuthToken{})
	resources.Registry().Register(&l8api.AuthUser{})
	clientConfig := &client.RestClientConfig{
		Host:          protocol.MachineIP,
		Port:          8080,
		Https:         true,
		TokenRequired: true,
		CertFileName:  "test.crt",
		Prefix:        prefix,
		AuthInfo: &client.RestAuthInfo{
			NeedAuth:   true,
			BodyType:   "AuthUser",
			UserField:  "User",
			PassField:  "Pass",
			RespType:   "AuthToken",
			TokenField: "Token",
			AuthPath:   "/auth",
		},
	}

	resources.Registry().Register(&l8api.AuthToken{})
	restClient, err := client.NewRestClient(clientConfig, resources)
	if err != nil {
		Log.Fail(t, err)
		return nil, false
	}
	resources.Registry().Register(pb)
	return restClient, true
}

func PushPlugin(nic ifs.IVNic, name string) error {
	data, err := os.ReadFile(name)
	if err != nil {
		return err
	}
	pb := &l8web.L8Plugin{
		Data: base64.StdEncoding.EncodeToString(data),
	}
	return plugins.LoadPlugin(pb, nic)
}
