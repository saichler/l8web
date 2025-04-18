package tests

import (
	. "github.com/saichler/l8test/go/infra/t_resources"
	"github.com/saichler/l8test/go/infra/t_servicepoints"
	"github.com/saichler/l8web/go/web/client"
	"github.com/saichler/l8web/go/web/server"
	"github.com/saichler/layer8/go/overlay/protocol"
	"github.com/saichler/types/go/testtypes"
	"net/http"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	setup()
	m.Run()
	tear()
}

func TestRestServer(t *testing.T) {
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
		return
	}

	snic := topo.VnicByVnetNum(3, 1)
	h := server.NewServicePointHandler(t_servicepoints.ServiceName, 0, snic)
	pb := &testtypes.TestProto{}
	h.AddMethodType(http.MethodPost, pb)

	srv.AddServicePointHandler(h)

	go srv.Start()
	time.Sleep(time.Second)

	cnic := topo.VnicByVnetNum(1, 2)

	clientConfig := &client.RestClientConfig{
		Host:         protocol.MachineIP,
		Port:         8080,
		Https:        true,
		CertFileName: "test.crt",
		Prefix:       "/test/",
	}
	clt, err := client.NewRestClient(clientConfig, cnic.Resources())
	if err != nil {
		Log.Fail(t, err)
		return
	}

	pb = &testtypes.TestProto{MyString: "Hello"}
	resp, err := clt.POST("0/"+t_servicepoints.ServiceName, "TestProto", "", "", pb)
	if err != nil {
		Log.Fail(t, err)
		return
	}
	if pb.MyString != resp.(*testtypes.TestProto).MyString {
		Log.Fail(t, "Expected the same object")
		return
	}
}
