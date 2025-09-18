package tests

import (
	"encoding/base64"
	"os"
	"reflect"
	"testing"
	"time"

	. "github.com/saichler/l8test/go/infra/t_resources"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types/l8health"
	"github.com/saichler/l8types/go/types/l8web"
	"github.com/saichler/l8web/go/web/client"
	"github.com/saichler/l8web/go/web/server"
	"github.com/saichler/layer8/go/overlay/plugins"
	"github.com/saichler/layer8/go/overlay/protocol"
	vnet2 "github.com/saichler/layer8/go/overlay/vnet"
	"github.com/saichler/layer8/go/overlay/vnic"
	"google.golang.org/protobuf/proto"
)

const (
	VNET_PORT = 28000
)

func TestMain(m *testing.M) {
	setup()
	m.Run()
	tear()
}

func TestRestServer(t *testing.T) {
	resources, _ := CreateResources(28000, 0, ifs.Info_Level)
	vnet := vnet2.NewVNet(resources)
	vnet.Start()
	time.Sleep(time.Second)

	webNic, svr, ok := createWebServer(t)
	if !ok {
		return
	}

	serviceNic, ok := createServiceNic(t)
	if !ok {
		return
	}

	defer func() {
		webNic.Shutdown()
		serviceNic.Shutdown()
		vnet.Shutdown()
		svr.Stop()
	}()

	info, err := serviceNic.Resources().Registry().Info("TestProtoList")
	if err != nil {
		Log.Fail(t, err)
		return
	}
	pbList, _ := info.NewInstance()

	info, _ = serviceNic.Resources().Registry().Info("TestProto")
	pb, _ := info.NewInstance()

	restClient, ok := createRestClient(t, pbList)
	if !ok {
		return
	}

	v := reflect.ValueOf(pb)
	field := v.Elem().FieldByName("MyString")
	field.Set(reflect.ValueOf("Hello"))

	resp, err := restClient.POST("0/Tests", "TestProtoList", "", "", pb.(proto.Message))
	if err != nil {
		Log.Fail(t, err)
		return
	}

	v = reflect.ValueOf(resp)
	field = v.Elem().FieldByName("List")
	field = field.Index(0).Elem().FieldByName("MyString")
	if field.String() != "Hello" {
		Log.Fail(t, "Expected the same object")
		return
	}
}

func TestRestServer2(t *testing.T) {
	resources, _ := CreateResources(28000, 0, ifs.Info_Level)
	vnet := vnet2.NewVNet(resources)
	vnet.Start()
	time.Sleep(time.Second)

	serviceNic, ok := createServiceNic(t)
	if !ok {
		return
	}

	info, err := serviceNic.Resources().Registry().Info("TestProtoList")
	if err != nil {
		Log.Fail(t, err)
		return
	}
	pbList, _ := info.NewInstance()

	info, _ = serviceNic.Resources().Registry().Info("TestProto")
	pb, _ := info.NewInstance()

	webNic, svr, ok := createWebServer(t)
	if !ok {
		return
	}

	defer func() {
		webNic.Shutdown()
		serviceNic.Shutdown()
		vnet.Shutdown()
		svr.Stop()
	}()

	time.Sleep(time.Second * 3)

	restClient, ok := createRestClient(t, pbList)
	if !ok {
		return
	}

	v := reflect.ValueOf(pb)
	field := v.Elem().FieldByName("MyString")
	field.Set(reflect.ValueOf("Hello"))

	resp, err := restClient.POST("0/Tests", "TestProtoList", "", "", pb.(proto.Message))
	if err != nil {
		Log.Fail(t, err)
		return
	}

	v = reflect.ValueOf(resp)
	field = v.Elem().FieldByName("List")
	field = field.Index(0).Elem().FieldByName("MyString")
	if field.String() != "Hello" {
		Log.Fail(t, "Expected the same object")
		return
	}
}

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
		Authentication: false,
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
	resources, _ := CreateResources(VNET_PORT, 3, ifs.Info_Level)
	clientConfig := &client.RestClientConfig{
		Host:         protocol.MachineIP,
		Port:         8080,
		Https:        true,
		CertFileName: "test.crt",
		Prefix:       "/test/",
	}
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
