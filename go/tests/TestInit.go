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

// Package tests provides test infrastructure and test cases for the l8web package.
// It includes utilities for creating test servers, clients, and VNets, as well as
// test cases for REST server functionality, authentication, and web service integration.

package tests

import (
	. "github.com/saichler/l8test/go/infra/t_resources"
	. "github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8bus/go/overlay/protocol"
)

//var topo *TestTopology

func init() {
	Log.SetLogLevel(Trace_Level)
}

func setup() {
	protocol.Discovery_Enabled = false
	setupTopology()
}

func tear() {
	shutdownTopology()
}

func reset(name string) {
	Log.Info("*** ", name, " end ***")
	//topo.ResetHandlers()
}

func setupTopology() {
	//topo = NewTestTopology(4, []int{20000, 30000, 40000}, Trace_Level)
}

func shutdownTopology() {
	//topo.Shutdown()
}
