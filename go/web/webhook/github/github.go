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

// Package github provides a webhook.Provider implementation for GitHub webhooks.
// It handles event type detection via the X-GitHub-Event header and HMAC-SHA256
// signature verification via the X-Hub-Signature-256 header.
package github

import (
	"encoding/json"
	"net/http"

	"github.com/saichler/l8web/go/web/webhook"
)

// Provider implements webhook.Provider for GitHub webhooks.
type Provider struct{}

// EventType returns the value of the X-GitHub-Event header.
func (p *Provider) EventType(r *http.Request) string {
	return r.Header.Get("X-GitHub-Event")
}

// VerifySignature checks the X-Hub-Signature-256 header against the payload
// using HMAC-SHA256 with the given secret.
func (p *Provider) VerifySignature(payload []byte, r *http.Request, secret string) bool {
	sig := r.Header.Get("X-Hub-Signature-256")
	return webhook.VerifyHMACSHA256(payload, sig, secret)
}

// PushEvent represents a GitHub push webhook payload.
type PushEvent struct {
	Ref     string   `json:"ref"`
	Commits []Commit `json:"commits"`
	Repo    Repo     `json:"repository"`
}

// PullRequestEvent represents a GitHub pull_request webhook payload.
type PullRequestEvent struct {
	Action string      `json:"action"`
	PR     PullRequest `json:"pull_request"`
	Repo   Repo        `json:"repository"`
}

// PullRequest contains pull request details from a GitHub webhook event.
type PullRequest struct {
	Title       string `json:"title"`
	Body        string `json:"body"`
	Merged      bool   `json:"merged"`
	MergeCommit string `json:"merge_commit_sha"`
	HTMLURL     string `json:"html_url"`
}

// Commit contains commit details from a GitHub push event.
type Commit struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

// Repo contains repository details from a GitHub webhook event.
type Repo struct {
	CloneURL string `json:"clone_url"`
	HTMLURL  string `json:"html_url"`
	FullName string `json:"full_name"`
}

// RepoURL extracts the repository HTML URL from a raw JSON payload
// without fully unmarshaling the event. Falls back to clone_url if
// html_url is empty.
func RepoURL(payload []byte) string {
	var p struct {
		Repo struct {
			CloneURL string `json:"clone_url"`
			HTMLURL  string `json:"html_url"`
		} `json:"repository"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return ""
	}
	if p.Repo.HTMLURL != "" {
		return p.Repo.HTMLURL
	}
	return p.Repo.CloneURL
}
