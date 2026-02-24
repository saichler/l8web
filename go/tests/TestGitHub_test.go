package tests

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/saichler/l8web/go/web/webhook/github"
)

func TestGitHubProvider_EventType(t *testing.T) {
	p := &github.Provider{}
	req := httptest.NewRequest(http.MethodPost, "/webhook", nil)
	req.Header.Set("X-GitHub-Event", "push")

	if et := p.EventType(req); et != "push" {
		t.Fatalf("expected 'push', got '%s'", et)
	}
}

func TestGitHubProvider_EventTypeMissing(t *testing.T) {
	p := &github.Provider{}
	req := httptest.NewRequest(http.MethodPost, "/webhook", nil)

	if et := p.EventType(req); et != "" {
		t.Fatalf("expected empty string, got '%s'", et)
	}
}

func TestGitHubProvider_VerifySignature_Valid(t *testing.T) {
	p := &github.Provider{}
	payload := []byte(`{"ref":"refs/heads/main"}`)
	secret := "webhook-secret"

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	sig := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	req := httptest.NewRequest(http.MethodPost, "/webhook", nil)
	req.Header.Set("X-Hub-Signature-256", sig)

	if !p.VerifySignature(payload, req, secret) {
		t.Fatal("expected valid signature to return true")
	}
}

func TestGitHubProvider_VerifySignature_Invalid(t *testing.T) {
	p := &github.Provider{}
	payload := []byte(`{"ref":"refs/heads/main"}`)

	req := httptest.NewRequest(http.MethodPost, "/webhook", nil)
	req.Header.Set("X-Hub-Signature-256", "sha256=0000000000000000000000000000000000000000000000000000000000000000")

	if p.VerifySignature(payload, req, "my-secret") {
		t.Fatal("expected invalid signature to return false")
	}
}

func TestRepoURL(t *testing.T) {
	payload := []byte(`{
		"ref": "refs/heads/main",
		"repository": {
			"html_url": "https://github.com/saichler/l8bugs",
			"clone_url": "https://github.com/saichler/l8bugs.git",
			"full_name": "saichler/l8bugs"
		}
	}`)

	url := github.RepoURL(payload)
	if url != "https://github.com/saichler/l8bugs" {
		t.Fatalf("expected 'https://github.com/saichler/l8bugs', got '%s'", url)
	}
}

func TestRepoURL_FallbackToClone(t *testing.T) {
	payload := []byte(`{
		"repository": {
			"clone_url": "https://github.com/saichler/l8bugs.git"
		}
	}`)

	url := github.RepoURL(payload)
	if url != "https://github.com/saichler/l8bugs.git" {
		t.Fatalf("expected clone URL fallback, got '%s'", url)
	}
}

func TestRepoURL_InvalidJSON(t *testing.T) {
	url := github.RepoURL([]byte("not json"))
	if url != "" {
		t.Fatalf("expected empty string for invalid JSON, got '%s'", url)
	}
}
