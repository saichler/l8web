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

type ProxyConfig struct {
	Listeners []ListenerConfig
}

type ListenerConfig struct {
	ListenPort string
	Routes     []RouteConfig
}

type RouteConfig struct {
	Domains    []string
	TargetPort string
	CertFile   string
	KeyFile    string
}

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

func Run() {
	proxy := NewReverseProxy()
	if err := proxy.Start(); err != nil {
		log.Fatal("Failed to start proxy:", err)
	}
}
