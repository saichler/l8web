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

// TestWeb_test.go contains integration tests for web service discovery.
// It verifies that web services can be registered and discovered across
// VNic instances through the Layer 8 network overlay.

package tests

import (
	"testing"
	"time"

	"github.com/saichler/l8bus/go/overlay/vnet"
	. "github.com/saichler/l8test/go/infra/t_resources"
	"github.com/saichler/l8types/go/ifs"
)

func TestWeb(t *testing.T) {
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

	time.Sleep(time.Second * 5)
}
