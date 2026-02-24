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

// Package gitlab provides a webhook.Provider implementation for GitLab webhooks.
// It handles event type detection via the X-Gitlab-Event header and secret token
// verification via the X-Gitlab-Token header (plain string comparison).
package gitlab

import (
	"encoding/json"
	"net/http"
)

// Provider implements webhook.Provider for GitLab webhooks.
type Provider struct{}

// EventType returns the value of the X-Gitlab-Event header.
func (p *Provider) EventType(r *http.Request) string {
	return r.Header.Get("X-Gitlab-Event")
}

// VerifySignature checks the X-Gitlab-Token header against the given secret
// using plain string comparison (GitLab does not use HMAC).
func (p *Provider) VerifySignature(payload []byte, r *http.Request, secret string) bool {
	token := r.Header.Get("X-Gitlab-Token")
	return token == secret
}

// PushEvent represents a GitLab push webhook payload.
type PushEvent struct {
	Ref     string   `json:"ref"`
	Commits []Commit `json:"commits"`
	Project Project  `json:"project"`
}

// MergeRequestEvent represents a GitLab merge_request webhook payload.
type MergeRequestEvent struct {
	ObjectAttrs MergeRequestAttrs `json:"object_attributes"`
	Project     Project           `json:"project"`
}

// MergeRequestAttrs contains merge request details from a GitLab webhook event.
type MergeRequestAttrs struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	State       string `json:"state"`
	MergeCommit string `json:"merge_commit_sha"`
	URL         string `json:"url"`
	Action      string `json:"action"`
}

// Commit contains commit details from a GitLab push event.
type Commit struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

// Project contains project details from a GitLab webhook event.
type Project struct {
	WebURL  string `json:"web_url"`
	HTTPURL string `json:"http_url"`
	Name    string `json:"name"`
	PathNS  string `json:"path_with_namespace"`
}

// RepoURL extracts the project web_url from a raw JSON payload
// without fully unmarshaling the event. Falls back to http_url if
// web_url is empty.
func RepoURL(payload []byte) string {
	var p struct {
		Project struct {
			WebURL  string `json:"web_url"`
			HTTPURL string `json:"http_url"`
		} `json:"project"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return ""
	}
	if p.Project.WebURL != "" {
		return p.Project.WebURL
	}
	return p.Project.HTTPURL
}
