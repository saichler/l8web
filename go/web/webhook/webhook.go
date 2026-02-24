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

// Package webhook provides a generic webhook HTTP handler for the Layer 8 framework.
// It handles POST-only enforcement, body reading, signature verification, and event
// dispatching. Use a Provider implementation (e.g., github.Provider) to plug in
// VCS-specific behavior.
package webhook

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

// Provider detects event type and verifies signatures for a webhook source.
type Provider interface {
	// EventType extracts the event type string from the HTTP request
	// (e.g., reads X-GitHub-Event header for GitHub).
	EventType(r *http.Request) string
	// VerifySignature checks the request's cryptographic signature
	// against the given secret. Returns true if valid or if secret is empty.
	VerifySignature(payload []byte, r *http.Request, secret string) bool
}

// EventHandler processes a webhook event. Returns an HTTP status code.
type EventHandler func(eventType string, payload []byte) int

// SecretFunc looks up the webhook secret for a request.
// Receives the raw payload so implementations can extract repo info.
// Return "" to skip signature verification.
type SecretFunc func(payload []byte) string

// NewHandler creates an http.Handler for a webhook endpoint.
// It enforces POST-only, reads the body, delegates event type detection
// and signature verification to the Provider, then calls the EventHandler.
func NewHandler(provider Provider, handler EventHandler, secretFn SecretFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusBadRequest)
			return
		}

		eventType := provider.EventType(r)
		if eventType == "" {
			http.Error(w, "missing event type", http.StatusBadRequest)
			return
		}

		secret := secretFn(body)
		if secret != "" {
			if !provider.VerifySignature(body, r, secret) {
				http.Error(w, "invalid signature", http.StatusForbidden)
				return
			}
		}

		status := handler(eventType, body)
		if status >= 400 {
			fmt.Fprintf(os.Stderr, "[webhook] handler returned %d for event %s\n", status, eventType)
		}
		w.WriteHeader(status)
	})
}
