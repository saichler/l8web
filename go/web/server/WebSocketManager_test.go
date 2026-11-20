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

// OnNotification delivers only to the targeted AaaIds and drops an
// untargeted notification instead of broadcasting it to every connection
// (probler/plans/websocket-page-scoped-notifications.md Phase 4).

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types/l8notify"
	"github.com/saichler/l8utils/go/utils/logger"
	"github.com/saichler/l8utils/go/utils/resources"
)

// logVnic provides resources (for the logger); every other IVNic method is
// unused here (the embedded nil interface would panic).
type logVnic struct {
	ifs.IVNic
	res ifs.IResources
}

func (this *logVnic) Resources() ifs.IResources { return this.res }

// newTestManager returns a manager with one live websocket connection per
// aaaId, and the client end of each.
func newTestManager(t *testing.T, aaaIds ...string) (*WebSocketManager, map[string]*websocket.Conn) {
	t.Helper()
	res := resources.NewResources(logger.NewLoggerDirectImpl(&logger.FmtLogMethod{}))
	m := NewWebSocketManager(&logVnic{res: res})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Error(err)
			return
		}
		m.mu.Lock()
		m.connections[r.URL.Query().Get("aaa")] = &wsConn{conn: conn}
		m.mu.Unlock()
	}))
	t.Cleanup(srv.Close)

	clients := make(map[string]*websocket.Conn)
	for _, id := range aaaIds {
		c, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(srv.URL, "http")+"/?aaa="+id, nil)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { c.Close() })
		clients[id] = c
	}
	deadline := time.Now().Add(2 * time.Second)
	for m.ConnectionCount() < len(aaaIds) {
		if time.Now().After(deadline) {
			t.Fatalf("expected %d connections, got %d", len(aaaIds), m.ConnectionCount())
		}
		time.Sleep(5 * time.Millisecond)
	}
	return m, clients
}

// readMsg returns the next message on c, or nil if none arrives in time.
func readMsg(c *websocket.Conn, wait time.Duration) map[string]interface{} {
	c.SetReadDeadline(time.Now().Add(wait))
	_, data, err := c.ReadMessage()
	if err != nil {
		return nil
	}
	msg := map[string]interface{}{}
	if json.Unmarshal(data, &msg) != nil {
		return nil
	}
	return msg
}

func TestOnNotificationDropsUntargeted(t *testing.T) {
	m, clients := newTestManager(t, "aaa-1", "aaa-2")

	m.OnNotification(&l8notify.L8NotificationSet{Type: l8notify.L8NotificationType_Patch,
		ModelType: "NetworkDevice", ModelKey: "dev-1", ServiceName: "NCache"})

	for id, c := range clients {
		if msg := readMsg(c, 300*time.Millisecond); msg != nil {
			t.Errorf("%s: expected no message for an untargeted notification, got %v", id, msg)
		}
	}
}

func TestOnNotificationDeliversOnlyToTargets(t *testing.T) {
	m, clients := newTestManager(t, "aaa-1", "aaa-2")

	m.OnNotification(&l8notify.L8NotificationSet{Type: l8notify.L8NotificationType_Patch,
		ModelType: "NetworkDevice", ModelKey: "dev-7", AaaIds: map[string]bool{"aaa-1": true}})

	msg := readMsg(clients["aaa-1"], 2*time.Second)
	if msg == nil {
		t.Fatal("aaa-1: expected the targeted notification")
	}
	if msg["action"] != "update" || msg["primaryKey"] != "dev-7" || msg["modelType"] != "NetworkDevice" {
		t.Errorf("aaa-1: unexpected message %v", msg)
	}
	if msg := readMsg(clients["aaa-2"], 300*time.Millisecond); msg != nil {
		t.Errorf("aaa-2: expected no message, got %v", msg)
	}
}
