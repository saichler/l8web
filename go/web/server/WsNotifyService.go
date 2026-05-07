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
	"fmt"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types/l8notify"
)

const (
	WsNotifyServiceName = "websock"
	WsNotifyServiceArea = byte(0)
)

// WsNotifyService is a stateless service that receives client-facing change
// notifications via L8Bus multicast and forwards them to WebSocket clients.
type WsNotifyService struct {
	wsManager *WebSocketManager
}

func NewWsNotifyService(wsManager *WebSocketManager) *WsNotifyService {
	return &WsNotifyService{wsManager: wsManager}
}

func (this *WsNotifyService) Activate(sla *ifs.ServiceLevelAgreement, vnic ifs.IVNic) error {
	if len(sla.Args()) > 0 {
		this.wsManager = sla.Args()[0].(*WebSocketManager)
	}
	fmt.Printf("[WS-DEBUG-3] WsNotifyService.Activate called: name=%s area=%d wsManager=%v\n", sla.ServiceName(), sla.ServiceArea(), this.wsManager != nil)
	return nil
}

func (this *WsNotifyService) DeActivate() error {
	return nil
}

func (this *WsNotifyService) Post(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	fmt.Println("[WS-DEBUG-3] WsNotifyService.Post called")
	this.handleNotification(pb)
	return nil
}

func (this *WsNotifyService) Put(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	fmt.Println("[WS-DEBUG-3] WsNotifyService.Put called")
	this.handleNotification(pb)
	return nil
}

func (this *WsNotifyService) Patch(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	fmt.Println("[WS-DEBUG-3] WsNotifyService.Patch called")
	this.handleNotification(pb)
	return nil
}

func (this *WsNotifyService) Delete(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	fmt.Println("[WS-DEBUG-3] WsNotifyService.Delete called")
	this.handleNotification(pb)
	return nil
}

func (this *WsNotifyService) Get(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	return nil
}

func (this *WsNotifyService) Failed(pb ifs.IElements, vnic ifs.IVNic, msg *ifs.Message) ifs.IElements {
	return nil
}

func (this *WsNotifyService) TransactionConfig() ifs.ITransactionConfig {
	return nil
}

func (this *WsNotifyService) WebService() ifs.IWebService {
	return nil
}

func (this *WsNotifyService) handleNotification(pb ifs.IElements) {
	fmt.Printf("[WS-DEBUG-3] handleNotification: wsManager=%v\n", this.wsManager != nil)
	if this.wsManager == nil {
		fmt.Println("[WS-DEBUG-3] wsManager is nil, returning")
		return
	}
	elem := pb.Element()
	fmt.Printf("[WS-DEBUG-3] element type=%T\n", elem)
	n, ok := elem.(*l8notify.L8NotificationSet)
	if !ok || n == nil {
		fmt.Printf("[WS-DEBUG-3] type assertion failed: ok=%v n=%v\n", ok, n)
		return
	}
	fmt.Printf("[WS-DEBUG-3] forwarding to wsManager: model=%s key=%s type=%v\n", n.ModelType, n.ModelKey, n.Type)
	this.wsManager.OnNotification(n)
}
