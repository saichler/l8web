package tests

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/saichler/l8web/go/web/webhook"
)

// mockProvider implements webhook.Provider for testing.
type mockProvider struct {
	eventType    string
	verifyResult bool
	verifyCalled bool
}

func (m *mockProvider) EventType(r *http.Request) string {
	return m.eventType
}

func (m *mockProvider) VerifySignature(payload []byte, r *http.Request, secret string) bool {
	m.verifyCalled = true
	return m.verifyResult
}

func TestNewHandler_PostOnly(t *testing.T) {
	provider := &mockProvider{eventType: "push"}
	handler := webhook.NewHandler(provider, func(string, []byte) int { return 200 }, func([]byte) string { return "" })

	req := httptest.NewRequest(http.MethodGet, "/webhook", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestNewHandler_EmptyEventType(t *testing.T) {
	provider := &mockProvider{eventType: ""}
	handler := webhook.NewHandler(provider, func(string, []byte) int { return 200 }, func([]byte) string { return "" })

	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader("{}"))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestNewHandler_SignatureFailure(t *testing.T) {
	provider := &mockProvider{eventType: "push", verifyResult: false}
	handler := webhook.NewHandler(provider, func(string, []byte) int { return 200 }, func([]byte) string { return "my-secret" })

	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader("{}"))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
	if !provider.verifyCalled {
		t.Fatal("expected VerifySignature to be called")
	}
}

func TestNewHandler_SignatureSkipped(t *testing.T) {
	provider := &mockProvider{eventType: "push", verifyResult: false}
	var called bool
	handler := webhook.NewHandler(provider, func(et string, p []byte) int {
		called = true
		return 200
	}, func([]byte) string { return "" })

	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader("{}"))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if provider.verifyCalled {
		t.Fatal("expected VerifySignature NOT to be called when secret is empty")
	}
	if !called {
		t.Fatal("expected EventHandler to be called")
	}
}

func TestNewHandler_Success(t *testing.T) {
	provider := &mockProvider{eventType: "push", verifyResult: true}
	var receivedEvent string
	var receivedPayload []byte
	handler := webhook.NewHandler(provider, func(et string, p []byte) int {
		receivedEvent = et
		receivedPayload = p
		return 201
	}, func([]byte) string { return "secret" })

	body := `{"ref":"refs/heads/main"}`
	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != 201 {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	if receivedEvent != "push" {
		t.Fatalf("expected event 'push', got '%s'", receivedEvent)
	}
	if string(receivedPayload) != body {
		t.Fatalf("expected payload '%s', got '%s'", body, string(receivedPayload))
	}
}
