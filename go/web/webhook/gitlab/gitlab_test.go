package gitlab

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestGitLabProvider_EventType(t *testing.T) {
	p := &Provider{}
	r, _ := http.NewRequest("POST", "/webhook", nil)
	r.Header.Set("X-Gitlab-Event", "Push Hook")
	if got := p.EventType(r); got != "Push Hook" {
		t.Fatalf("expected 'Push Hook', got %q", got)
	}
}

func TestGitLabProvider_EventTypeMissing(t *testing.T) {
	p := &Provider{}
	r, _ := http.NewRequest("POST", "/webhook", nil)
	if got := p.EventType(r); got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
}

func TestGitLabProvider_VerifySignature_Valid(t *testing.T) {
	p := &Provider{}
	r, _ := http.NewRequest("POST", "/webhook", nil)
	r.Header.Set("X-Gitlab-Token", "my-secret-token")
	if !p.VerifySignature(nil, r, "my-secret-token") {
		t.Fatal("expected valid signature")
	}
}

func TestGitLabProvider_VerifySignature_Invalid(t *testing.T) {
	p := &Provider{}
	r, _ := http.NewRequest("POST", "/webhook", nil)
	r.Header.Set("X-Gitlab-Token", "wrong-token")
	if p.VerifySignature(nil, r, "my-secret-token") {
		t.Fatal("expected invalid signature")
	}
}

func TestGitLabProvider_VerifySignature_Missing(t *testing.T) {
	p := &Provider{}
	r, _ := http.NewRequest("POST", "/webhook", nil)
	if p.VerifySignature(nil, r, "my-secret-token") {
		t.Fatal("expected invalid signature when header missing")
	}
}

func TestRepoURL(t *testing.T) {
	ev := map[string]interface{}{
		"ref": "refs/heads/main",
		"project": map[string]interface{}{
			"web_url":  "https://gitlab.com/test/repo",
			"http_url": "https://gitlab.com/test/repo.git",
		},
	}
	data, _ := json.Marshal(ev)
	got := RepoURL(data)
	if got != "https://gitlab.com/test/repo" {
		t.Fatalf("expected 'https://gitlab.com/test/repo', got %q", got)
	}
}

func TestRepoURL_FallbackToHTTP(t *testing.T) {
	ev := map[string]interface{}{
		"project": map[string]interface{}{
			"http_url": "https://gitlab.com/test/repo.git",
		},
	}
	data, _ := json.Marshal(ev)
	got := RepoURL(data)
	if got != "https://gitlab.com/test/repo.git" {
		t.Fatalf("expected fallback to http_url, got %q", got)
	}
}

func TestRepoURL_InvalidJSON(t *testing.T) {
	got := RepoURL([]byte("not json"))
	if got != "" {
		t.Fatalf("expected empty string for invalid JSON, got %q", got)
	}
}
