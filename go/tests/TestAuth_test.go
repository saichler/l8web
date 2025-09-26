package tests

import (
	"testing"
	"time"

	vnet2 "github.com/saichler/l8bus/go/overlay/vnet"
	. "github.com/saichler/l8test/go/infra/t_resources"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types/l8api"
)

func TestAuth(t *testing.T) {
	resources, _ := CreateResources(28000, 0, ifs.Info_Level)
	vnet := vnet2.NewVNet(resources)
	vnet.Start()
	time.Sleep(time.Second)

	/*
		serviceNic, ok := createServiceNic(t)
		if !ok {
			return
		}

		info, err := serviceNic.Resources().Registry().Info("TestProtoList")
		if err != nil {
			Log.Fail(t, err)
			return
		}

		//pbList, _ := info.NewInstance()

		info, _ = serviceNic.Resources().Registry().Info("TestProto")
		pb, _ := info.NewInstance()
	*/

	webNic, svr, ok := createWebServer(t)
	if !ok {
		return
	}

	defer func() {
		webNic.Shutdown()
		//serviceNic.Shutdown()
		vnet.Shutdown()
		svr.Stop()
	}()

	user := &l8api.AuthUser{User: "admin", Pass: "admin"}

	restClient, ok := createRestClient2(t, user, "/")
	if !ok {
		return
	}

	resp, err := restClient.POST("auth", "AuthToken", "", "", user)
	if err != nil {
		Log.Fail(t, err)
		return
	}

	authToken := resp.(*l8api.AuthToken)
	if authToken.Error != "" {
		Log.Fail(t, authToken.Error)
		return
	}

	user.Pass = "not the pass"
	resp, err = restClient.POST("auth", "AuthToken", "", "", user)
	if err == nil {
		Log.Fail(t, "Expected auth failure")
		return
	}
}
