// © 2025 Sharon Aicler (saichler@gmail.com)
//
// Layer 8 Ecosystem is licensed under the Apache License, Version 2.0.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"sync"

	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types/l8api"
)

// registration is one live-notification registration a client component
// holds with a service, as forwarded by this web server.
type registration struct {
	serviceName string
	serviceArea byte
	registerId  string
	text        string
}

// registrations records, per user (AaaId), the registrations this web server
// forwarded, so that when the user's websocket closes it can unregister them
// with their services: registrations have no time-based expiry, so a closed
// tab would otherwise leave them behind for good.
type registrations struct {
	mu   sync.Mutex
	byId map[string]map[string]*registration
}

func newRegistrations() *registrations {
	return &registrations{byId: make(map[string]map[string]*registration)}
}

var trackedRegistrations = newRegistrations()

// track records or forgets the registration a forwarded query carries.
func (this *registrations) track(aaaId, serviceName string, serviceArea byte, q *l8api.L8Query) {
	if aaaId == "" || q.RegisterId == "" {
		return
	}
	key := serviceKey(serviceName, serviceArea) + "/" + q.RegisterId
	this.mu.Lock()
	defer this.mu.Unlock()
	switch q.Register {
	case l8api.L8Register_L8_REGISTER_REGISTER:
		m, ok := this.byId[aaaId]
		if !ok {
			m = make(map[string]*registration)
			this.byId[aaaId] = m
		}
		m[key] = &registration{serviceName: serviceName, serviceArea: serviceArea,
			registerId: q.RegisterId, text: q.Text}
	case l8api.L8Register_L8_REGISTER_UNREGISTER:
		if m, ok := this.byId[aaaId]; ok {
			delete(m, key)
			if len(m) == 0 {
				delete(this.byId, aaaId)
			}
		}
	}
}

// take removes and returns every registration recorded for aaaId.
func (this *registrations) take(aaaId string) []*registration {
	this.mu.Lock()
	defer this.mu.Unlock()
	m := this.byId[aaaId]
	delete(this.byId, aaaId)
	result := make([]*registration, 0, len(m))
	for _, r := range m {
		result = append(result, r)
	}
	return result
}

// forward routes a request to a service by the configured Target / Method,
// the one routing rule for both client requests and server-side unregisters.
func forward(vnic ifs.IVNic, serviceName string, serviceArea byte, action ifs.Action,
	body interface{}, timeout int, aaaId string) ifs.IElements {
	if Target != "" {
		return vnic.Request(Target, serviceName, serviceArea, action, body, timeout, aaaId)
	}
	if Method == ifs.M_Leader {
		return vnic.LeaderRequest(serviceName, serviceArea, action, body, timeout, aaaId)
	}
	if Method == ifs.M_Local {
		return vnic.LocalRequest(serviceName, serviceArea, action, body, timeout, aaaId)
	}
	return vnic.ProximityRequest(serviceName, serviceArea, action, body, timeout, aaaId)
}
