package tests

import (
	"reflect"
	"testing"
	"time"

	"github.com/saichler/l8bus/go/overlay/vnet"
	. "github.com/saichler/l8test/go/infra/t_resources"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8web/go/web/server"
	"google.golang.org/protobuf/proto"
)

func TestMain(m *testing.M) {
	setup()
	m.Run()
	tear()
}

func TestRestServer(t *testing.T) {
	resources, _ := CreateResources(28000, 0, ifs.Info_Level)
	vnet := vnet.NewVNet(resources)
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

	err = restClient.Auth("admin", "admin")
	if err != nil {
		Log.Fail(t, err.Error())
		return
	}

	v := reflect.ValueOf(pb)
	field := v.Elem().FieldByName("MyString")
	field.Set(reflect.ValueOf("Hello"))

	server.Target = serviceNic.Resources().SysConfig().LocalUuid

	time.Sleep(time.Second)

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
	vnet := vnet.NewVNet(resources)
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

	err = restClient.Auth("admin", "admin")
	if err != nil {
		Log.Fail(t, err.Error())
		return
	}

	v := reflect.ValueOf(pb)
	field := v.Elem().FieldByName("MyString")
	field.Set(reflect.ValueOf("Hello"))

	server.Target = serviceNic.Resources().SysConfig().LocalUuid

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
