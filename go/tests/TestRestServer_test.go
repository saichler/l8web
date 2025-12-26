/*
 * Copyright (c) 2025 Sharon Aicler (saichler@gmail.com)
 *
 * Layer 8 Ecosystem is licensed under the Apache License, Version 2.0.
 * You may obtain a copy of the License at:
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// TestRestServer_test.go contains integration tests for the REST server.
// It verifies end-to-end functionality including:
//   - Server creation and configuration
//   - Service registration via plugins
//   - Client authentication
//   - Protocol Buffer request/response handling
//   - Targeted routing to specific service instances

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
