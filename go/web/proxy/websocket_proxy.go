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

package proxy

import (
	"crypto/tls"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"sort"
	"strings"
	"sync"
)

func isWebSocketUpgrade(r *http.Request) bool {
	conn := strings.ToLower(r.Header.Get("Connection"))
	upgrade := strings.ToLower(r.Header.Get("Upgrade"))
	isWS := strings.Contains(conn, "upgrade") && upgrade == "websocket"
	log.Printf("[WS-DEBUG] isWebSocketUpgrade check: Connection=%q, Upgrade=%q, result=%v",
		r.Header.Get("Connection"), r.Header.Get("Upgrade"), isWS)
	return isWS
}

func logRequestHeaders(prefix string, r *http.Request) {
	log.Printf("[WS-DEBUG] %s: %s %s Host=%q RemoteAddr=%s", prefix, r.Method, r.URL.String(), r.Host, r.RemoteAddr)
	keys := make([]string, 0, len(r.Header))
	for k := range r.Header {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		log.Printf("[WS-DEBUG]   Header %s: %s", k, strings.Join(r.Header[k], ", "))
	}
}

func proxyWebSocket(w http.ResponseWriter, r *http.Request, backendHost string, backendPort string) {
	log.Printf("[WS-DEBUG] proxyWebSocket called: backendHost=%s backendPort=%s", backendHost, backendPort)
	logRequestHeaders("proxyWebSocket incoming request", r)

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		log.Printf("[WS-DEBUG] ERROR: ResponseWriter does not implement http.Hijacker")
		http.Error(w, "WebSocket hijack not supported", http.StatusInternalServerError)
		return
	}
	log.Printf("[WS-DEBUG] ResponseWriter supports Hijack")

	backendAddr := net.JoinHostPort(backendHost, backendPort)
	log.Printf("[WS-DEBUG] Dialing TLS to backend %s ...", backendAddr)
	backendConn, err := tls.Dial("tcp", backendAddr, &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		log.Printf("[WS-DEBUG] ERROR: TLS dial to backend %s failed: %v", backendAddr, err)
		http.Error(w, "Backend connection failed", http.StatusBadGateway)
		return
	}
	log.Printf("[WS-DEBUG] TLS connection to backend %s established (remote=%s)", backendAddr, backendConn.RemoteAddr())

	log.Printf("[WS-DEBUG] Writing original HTTP request to backend...")
	err = r.Write(backendConn)
	if err != nil {
		backendConn.Close()
		log.Printf("[WS-DEBUG] ERROR: writing request to backend failed: %v", err)
		http.Error(w, "Failed to forward request", http.StatusBadGateway)
		return
	}
	log.Printf("[WS-DEBUG] Request written to backend successfully")

	log.Printf("[WS-DEBUG] Hijacking client connection...")
	clientConn, clientBuf, err := hijacker.Hijack()
	if err != nil {
		backendConn.Close()
		log.Printf("[WS-DEBUG] ERROR: hijack failed: %v", err)
		return
	}
	log.Printf("[WS-DEBUG] Client connection hijacked successfully (local=%s, remote=%s)",
		clientConn.LocalAddr(), clientConn.RemoteAddr())

	if clientBuf.Reader.Buffered() > 0 {
		buffered := make([]byte, clientBuf.Reader.Buffered())
		n, _ := clientBuf.Read(buffered)
		if n > 0 {
			log.Printf("[WS-DEBUG] Flushing %d buffered bytes from client to backend", n)
			backendConn.Write(buffered[:n])
		}
	}

	log.Printf("[WS-DEBUG] WebSocket proxy tunnel established: client(%s) <-> backend(%s)", r.RemoteAddr, backendAddr)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		n, err := io.Copy(clientConn, backendConn)
		log.Printf("[WS-DEBUG] backend->client copy ended: %d bytes, err=%v", n, err)
		clientConn.Close()
	}()

	go func() {
		defer wg.Done()
		n, err := io.Copy(backendConn, clientConn)
		log.Printf("[WS-DEBUG] client->backend copy ended: %d bytes, err=%v", n, err)
		backendConn.Close()
	}()

	wg.Wait()
	log.Printf("[WS-DEBUG] WebSocket proxy tunnel closed: client(%s) <-> backend(%s)", r.RemoteAddr, backendAddr)
}

func makeHandler(domain string, hostname string, targetPort string, proxy *httputil.ReverseProxy) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[WS-DEBUG] makeHandler[%s]: %s %s", domain, r.Method, r.URL.String())
		if isWebSocketUpgrade(r) {
			log.Printf("[WS-DEBUG] makeHandler[%s]: detected WebSocket upgrade, proxying to %s:%s", domain, hostname, targetPort)
			proxyWebSocket(w, r, hostname, targetPort)
			return
		}
		log.Printf("Proxying request from %s to backend", domain)
		proxy.ServeHTTP(w, r)
	}
}
