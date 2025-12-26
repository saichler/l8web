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

// Package main provides the entry point for the Layer 8 reverse proxy service.
// It starts the SNI-based TLS reverse proxy with default configuration for
// multi-domain, multi-port SSL termination and routing.
//
// Environment Variables:
//   - NODE_IP: Backend host address (defaults to "localhost")
//
// Usage:
//
//	go build -o l8proxy main.go
//	sudo ./l8proxy  # requires root for port 443
//
// The proxy listens on ports 443, 14443, 9092, and 9094 by default.
package main

import (
	"github.com/saichler/l8web/go/web/proxy"
)

// main starts the Layer 8 reverse proxy with default configuration.
// It blocks until an error occurs.
func main() {
	proxy.Run()
}