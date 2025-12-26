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

// Package proxy provides an SNI-based TLS reverse proxy for the Layer 8 ecosystem.
// It supports multi-domain, multi-port routing with automatic TLS certificate selection
// based on the Server Name Indication (SNI) in the TLS handshake.
//
// Features:
//   - SNI-based certificate selection for multi-domain hosting
//   - Multi-port listening (443, 14443, 9092, 9094, etc.)
//   - Per-route SSL certificate configuration
//   - Environment-based backend host configuration (NODE_IP)
//   - Fallback domain matching for unmatched routes
//
// Default route configuration:
//   - Port 443: layer8vibe.dev->1443, probler.dev->2443, layer-8.dev->4443
//   - Port 14443: probler.dev->13443
//   - Port 9092: probler.dev->9093
//   - Port 9094: probler.dev->9095
package proxy

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

// ProxyConfig holds the complete configuration for the reverse proxy,
// including all listeners and their routing rules.
type ProxyConfig struct {
	Listeners []ListenerConfig // List of port listeners to start
}

// ListenerConfig defines a single port listener with its routing rules.
// Each listener can have multiple routes for different domains.
type ListenerConfig struct {
	ListenPort string        // Port to listen on (e.g., ":443", ":14443")
	Routes     []RouteConfig // Routing rules for this listener
}

// RouteConfig defines a single routing rule that maps domains to a backend port.
// Each route has its own SSL certificate for TLS termination.
type RouteConfig struct {
	Domains    []string // Domain names to match (e.g., ["www.example.com", "example.com"])
	TargetPort string   // Backend port to proxy to (e.g., "1443")
	CertFile   string   // Path to SSL certificate file
	KeyFile    string   // Path to SSL private key file
}

// NewReverseProxy creates a ProxyConfig with the default Layer 8 routing configuration.
// This includes listeners for ports 443, 14443, 9092, and 9094 with routes to
// layer8vibe.dev, probler.dev, and layer-8.dev domains.
func NewReverseProxy() *ProxyConfig {
	return &ProxyConfig{
		Listeners: []ListenerConfig{
			{
				ListenPort: ":443",
				Routes: []RouteConfig{
					{
						Domains:    []string{"www.layer8vibe.dev", "layer8vibe.dev"},
						TargetPort: "1443",
						CertFile:   "layer8vibe.dev/domain.cert.pem",
						KeyFile:    "layer8vibe.dev/private.key.pem",
					},
					{
						Domains:    []string{"www.probler.dev", "probler.dev"},
						TargetPort: "2443",
						CertFile:   "probler.dev/domain.cert.pem",
						KeyFile:    "probler.dev/private.key.pem",
					},
					{
						Domains:    []string{"www.layer-8.dev", "layer-8.dev"},
						TargetPort: "4443",
						CertFile:   "layer-8.dev/domain.cert.pem",
						KeyFile:    "layer-8.dev/private.key.pem",
					},
				},
			},
			{
				ListenPort: ":14443",
				Routes: []RouteConfig{
					{
						Domains:    []string{"www.probler.dev", "probler.dev"},
						TargetPort: "13443",
						CertFile:   "probler.dev/domain.cert.pem",
						KeyFile:    "probler.dev/private.key.pem",
					},
				},
			},
			{
				ListenPort: ":9092",
				Routes: []RouteConfig{
					{
						Domains:    []string{"www.probler.dev", "probler.dev"},
						TargetPort: "9093",
						CertFile:   "probler.dev/domain.cert.pem",
						KeyFile:    "probler.dev/private.key.pem",
					},
				},
			},
			{
				ListenPort: ":9094",
				Routes: []RouteConfig{
					{
						Domains:    []string{"www.probler.dev", "probler.dev"},
						TargetPort: "9095",
						CertFile:   "probler.dev/domain.cert.pem",
						KeyFile:    "probler.dev/private.key.pem",
					},
				},
			},
		},
	}
}

// Start begins all configured listeners in separate goroutines.
// It blocks until one of the listeners returns an error, then returns that error.
// Each listener runs in its own goroutine for concurrent multi-port operation.
func (pc *ProxyConfig) Start() error {
	errChan := make(chan error, len(pc.Listeners))

	for _, listener := range pc.Listeners {
		go func(listener ListenerConfig) {
			if err := pc.startListener(listener); err != nil {
				errChan <- err
			}
		}(listener)
	}

	// Wait for first error from any listener
	return <-errChan
}

// startListener initializes and starts a single port listener.
// It creates reverse proxy handlers for each route, sets up SNI-based certificate
// selection, and starts the HTTPS server. The backend host is determined by the
// NODE_IP environment variable (defaults to "localhost").
//
// The function sets up two types of handlers:
// 1. Domain-specific pattern handlers (e.g., "example.com/")
// 2. A fallback root handler ("/") that matches domains by Host header
func (pc *ProxyConfig) startListener(listener ListenerConfig) error {
	mux := http.NewServeMux()

	hostname := os.Getenv("NODE_IP")
	if hostname == "" {
		hostname = "localhost"
	}

	for _, route := range listener.Routes {
		targetURL, err := url.Parse(fmt.Sprintf("https://%s:%s", hostname, route.TargetPort))
		if err != nil {
			return fmt.Errorf("failed to parse target URL for port %s: %v", route.TargetPort, err)
		}

		proxy := httputil.NewSingleHostReverseProxy(targetURL)

		originalDirector := proxy.Director
		proxy.Director = func(req *http.Request) {
			originalDirector(req)
			req.Host = req.URL.Host
			req.URL.Scheme = "https"
		}

		proxy.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}

		for _, domain := range route.Domains {
			pattern := fmt.Sprintf("%s/", domain)
			mux.HandleFunc(pattern, func(domain string, proxy *httputil.ReverseProxy) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					log.Printf("Proxying request from %s to backend", domain)
					proxy.ServeHTTP(w, r)
				}
			}(domain, proxy))
		}
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		host := strings.ToLower(r.Host)

		for _, route := range listener.Routes {
			for _, domain := range route.Domains {
				// Strip port from host for comparison
				hostWithoutPort := strings.Split(host, ":")[0]
				if hostWithoutPort == domain || host == domain {
					hostname := os.Getenv("NODE_IP")
					if hostname == "" {
						hostname = "localhost"
					}
					targetURL, _ := url.Parse(fmt.Sprintf("https://%s:%s", hostname, route.TargetPort))
					proxy := httputil.NewSingleHostReverseProxy(targetURL)

					originalDirector := proxy.Director
					proxy.Director = func(req *http.Request) {
						originalDirector(req)
						req.Host = req.URL.Host
						req.URL.Scheme = "https"
					}

					proxy.Transport = &http.Transport{
						TLSClientConfig: &tls.Config{
							InsecureSkipVerify: true,
						},
					}

					log.Printf("Proxying request from %s to %s:%s", host, hostname, route.TargetPort)
					proxy.ServeHTTP(w, r)
					return
				}
			}
		}

		http.Error(w, "Unknown host", http.StatusBadGateway)
	})

	tlsConfig := &tls.Config{
		GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
			return pc.getCertificateForListener(info, listener)
		},
	}

	server := &http.Server{
		Addr:      listener.ListenPort,
		Handler:   mux,
		TLSConfig: tlsConfig,
	}

	log.Printf("Starting reverse proxy on port %s", listener.ListenPort)
	return server.ListenAndServeTLS("", "")
}

// getCertificateForListener implements SNI-based certificate selection.
// It searches the listener's routes for a matching domain and returns the
// corresponding certificate. If no match is found, it falls back to the
// first route's certificate (for domain aliases or misconfigured clients).
//
// This function is called during the TLS handshake via tls.Config.GetCertificate.
func (pc *ProxyConfig) getCertificateForListener(info *tls.ClientHelloInfo, listener ListenerConfig) (*tls.Certificate, error) {
	host := strings.ToLower(info.ServerName)

	for _, route := range listener.Routes {
		for _, domain := range route.Domains {
			if host == domain {
				cert, err := tls.LoadX509KeyPair(route.CertFile, route.KeyFile)
				if err != nil {
					log.Printf("Error loading certificate for %s: %v", domain, err)
					return nil, err
				}
				return &cert, nil
			}
		}
	}

	// Fallback to first route's certificate
	if len(listener.Routes) > 0 {
		cert, err := tls.LoadX509KeyPair(listener.Routes[0].CertFile, listener.Routes[0].KeyFile)
		if err != nil {
			return nil, err
		}
		return &cert, nil
	}

	return nil, fmt.Errorf("no certificate found for host: %s", host)
}

// Run creates a new reverse proxy with default configuration and starts it.
// This is the main entry point for running the proxy as a standalone service.
// It blocks until an error occurs and calls log.Fatal on failure.
func Run() {
	proxy := NewReverseProxy()
	if err := proxy.Start(); err != nil {
		log.Fatal("Failed to start proxy:", err)
	}
}
