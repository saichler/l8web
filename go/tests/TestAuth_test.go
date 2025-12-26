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

// TestAuth_test.go contains tests for the authentication endpoints.
// It verifies that the /auth endpoint correctly authenticates valid credentials
// and rejects invalid credentials.

package tests

import (
	"fmt"
	"testing"
	"time"

	vnet2 "github.com/saichler/l8bus/go/overlay/vnet"
	. "github.com/saichler/l8test/go/infra/t_resources"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types/l8api"
	"google.golang.org/protobuf/encoding/protojson"
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
	jsn, _ := protojson.Marshal(user)
	fmt.Println(string(jsn))
	restClient, ok := createRestClient2(t, user, "/")
	if !ok {
		return
	}

	resp, err := restClient.POST("auth", "AuthToken", "", "", user)
	if err != nil {
		Log.Fail(t, err)
		return
	}

	jsn, _ = protojson.Marshal(resp)
	fmt.Println(string(jsn))

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
	jsn, _ = protojson.Marshal(resp)
	fmt.Println(string(jsn))
}
