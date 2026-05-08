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
	"strings"
	"sync"
)

func isWebSocketUpgrade(r *http.Request) bool {
	conn := strings.ToLower(r.Header.Get("Connection"))
	upgrade := strings.ToLower(r.Header.Get("Upgrade"))
	return strings.Contains(conn, "upgrade") && upgrade == "websocket"
}

func proxyWebSocket(w http.ResponseWriter, r *http.Request, backendHost string, backendPort string) {
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "WebSocket hijack not supported", http.StatusInternalServerError)
		return
	}

	backendAddr := net.JoinHostPort(backendHost, backendPort)
	backendConn, err := tls.Dial("tcp", backendAddr, &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		log.Printf("WebSocket proxy: failed to connect to backend %s: %v", backendAddr, err)
		http.Error(w, "Backend connection failed", http.StatusBadGateway)
		return
	}

	err = r.Write(backendConn)
	if err != nil {
		backendConn.Close()
		log.Printf("WebSocket proxy: failed to write request to backend: %v", err)
		http.Error(w, "Failed to forward request", http.StatusBadGateway)
		return
	}

	clientConn, clientBuf, err := hijacker.Hijack()
	if err != nil {
		backendConn.Close()
		log.Printf("WebSocket proxy: hijack failed: %v", err)
		return
	}

	// Flush any buffered data from the client reader to the backend
	if clientBuf.Reader.Buffered() > 0 {
		buffered := make([]byte, clientBuf.Reader.Buffered())
		n, _ := clientBuf.Read(buffered)
		if n > 0 {
			backendConn.Write(buffered[:n])
		}
	}

	log.Printf("WebSocket proxy: connected %s -> %s", r.Host, backendAddr)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		io.Copy(clientConn, backendConn)
		clientConn.Close()
	}()

	go func() {
		defer wg.Done()
		io.Copy(backendConn, clientConn)
		backendConn.Close()
	}()

	wg.Wait()
}

func makeHandler(domain string, hostname string, targetPort string, proxy *httputil.ReverseProxy) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if isWebSocketUpgrade(r) {
			log.Printf("WebSocket upgrade from %s, proxying to %s:%s", domain, hostname, targetPort)
			proxyWebSocket(w, r, hostname, targetPort)
			return
		}
		log.Printf("Proxying request from %s to backend", domain)
		proxy.ServeHTTP(w, r)
	}
}
