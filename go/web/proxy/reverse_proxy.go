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
				Domains:    []string{"www.probler.dev:13443", "probler.dev:13443"},
				TargetPort: "13443",
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
	}
}

func (pc *ProxyConfig) Start() error {
	mux := http.NewServeMux()

	hostname := os.Getenv("NODE_IP")
	if hostname == "" {
		hostname = "localhost"
	}

	for _, route := range pc.Routes {
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

		for _, route := range pc.Routes {
			for _, domain := range route.Domains {
				if host == domain || host == domain+":443" {
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
		GetCertificate: pc.getCertificate,
	}

	server := &http.Server{
		Addr:      pc.ListenPort,
		Handler:   mux,
		TLSConfig: tlsConfig,
	}

	log.Printf("Starting reverse proxy on port %s", pc.ListenPort)
	return server.ListenAndServeTLS("", "")
}

func (pc *ProxyConfig) getCertificate(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
	host := strings.ToLower(info.ServerName)

	for _, route := range pc.Routes {
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

	if len(pc.Routes) > 0 {
		cert, err := tls.LoadX509KeyPair(pc.Routes[0].CertFile, pc.Routes[0].KeyFile)
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
