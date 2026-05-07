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
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types/l8notify"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type wsConn struct {
	conn *websocket.Conn
	wmu  sync.Mutex
}

func (c *wsConn) writeJSON(data []byte) error {
	c.wmu.Lock()
	defer c.wmu.Unlock()
	return c.conn.WriteMessage(websocket.TextMessage, data)
}

// WebSocketManager manages WebSocket connections keyed by AAAId (authenticated user identity).
type WebSocketManager struct {
	mu          sync.RWMutex
	connections map[string]*wsConn
	vnic        ifs.IVNic
}

func NewWebSocketManager(vnic ifs.IVNic) *WebSocketManager {
	return &WebSocketManager{
		connections: make(map[string]*wsConn),
		vnic:        vnic,
	}
}

// HandleUpgrade validates the bearer token, resolves the AAAId, and upgrades to a WebSocket connection.
func (this *WebSocketManager) HandleUpgrade(w http.ResponseWriter, r *http.Request) {
	fmt.Println("[WS-DEBUG-4] HandleUpgrade called")
	token := extractToken(r)
	if token == "" {
		fmt.Println("[WS-DEBUG-4] no token, returning 401")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	aaaId, ok := this.vnic.Resources().Security().ValidateToken(token, this.vnic)
	if !ok {
		fmt.Println("[WS-DEBUG-4] invalid token, returning 401")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	fmt.Printf("[WS-DEBUG-4] token valid, aaaId=%s\n", aaaId)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("[WS-DEBUG-4] WebSocket upgrade failed:", err)
		return
	}

	wc := &wsConn{conn: conn}

	this.mu.Lock()
	old, exists := this.connections[aaaId]
	if exists {
		old.conn.Close()
	}
	this.connections[aaaId] = wc
	this.mu.Unlock()

	fmt.Printf("[WS-DEBUG-4] WebSocket connected: aaaId=%s totalConns=%d\n", aaaId, this.ConnectionCount())
	go this.writePump(wc)
	go this.readPump(aaaId, wc)
}

// readPump reads from the connection to detect close and handle pings.
func (this *WebSocketManager) readPump(aaaId string, wc *wsConn) {
	defer this.Remove(aaaId)
	wc.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	wc.conn.SetPongHandler(func(string) error {
		wc.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	for {
		_, _, err := wc.conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

// writePump sends periodic pings to keep the connection alive.
func (this *WebSocketManager) writePump(wc *wsConn) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		wc.wmu.Lock()
		err := wc.conn.WriteMessage(websocket.PingMessage, nil)
		wc.wmu.Unlock()
		if err != nil {
			return
		}
	}
}

// Remove closes and removes the connection for the given AAAId.
func (this *WebSocketManager) Remove(aaaId string) {
	this.mu.Lock()
	wc, ok := this.connections[aaaId]
	if ok {
		wc.conn.Close()
		delete(this.connections, aaaId)
	}
	this.mu.Unlock()
}

// OnNotification serializes a notification and sends to subscribed clients.
func (this *WebSocketManager) OnNotification(notification *l8notify.L8NotificationSet) {
	fmt.Printf("[WS-DEBUG-5] OnNotification: model=%s key=%s type=%v connCount=%d\n", notification.ModelType, notification.ModelKey, notification.Type, this.ConnectionCount())
	action := ""
	switch notification.Type {
	case l8notify.L8NotificationType_Post:
		action = "add"
	case l8notify.L8NotificationType_Put, l8notify.L8NotificationType_Patch:
		action = "update"
	case l8notify.L8NotificationType_Delete:
		action = "delete"
	default:
		fmt.Printf("[WS-DEBUG-5] unknown notification type %v, returning\n", notification.Type)
		return
	}

	msg := map[string]interface{}{
		"action":     action,
		"modelType":  notification.ModelType,
		"primaryKey": notification.ModelKey,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		fmt.Printf("[WS-DEBUG-5] json marshal error: %v\n", err)
		return
	}
	fmt.Printf("[WS-DEBUG-5] sending to clients: %s\n", string(data))

	this.mu.RLock()
	defer this.mu.RUnlock()

	if len(notification.AaaIds) == 0 {
		fmt.Printf("[WS-DEBUG-5] broadcasting to all %d connections\n", len(this.connections))
		for aaaId, wc := range this.connections {
			if err := wc.writeJSON(data); err != nil {
				fmt.Printf("[WS-DEBUG-5] write error for %s: %v\n", aaaId, err)
				go this.Remove(aaaId)
			} else {
				fmt.Printf("[WS-DEBUG-5] sent to %s OK\n", aaaId)
			}
		}
		return
	}

	for aaaId := range notification.AaaIds {
		wc, ok := this.connections[aaaId]
		if !ok {
			continue
		}
		if err := wc.writeJSON(data); err != nil {
			go this.Remove(aaaId)
		}
	}
}

// ConnectionCount returns the number of active WebSocket connections.
func (this *WebSocketManager) ConnectionCount() int {
	this.mu.RLock()
	defer this.mu.RUnlock()
	return len(this.connections)
}
