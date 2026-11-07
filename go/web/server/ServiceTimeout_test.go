package server

import "testing"

func TestServiceTimeoutOverride(t *testing.T) {
	chat := &ServiceHandler{serviceName: "AgntChat", serviceArea: 120}
	other := &ServiceHandler{serviceName: "Employee", serviceArea: 30}
	sameNameOtherArea := &ServiceHandler{serviceName: "AgntChat", serviceArea: 121}

	if chat.timeout() != Timeout {
		t.Fatalf("no override: got %d, want the default %d", chat.timeout(), Timeout)
	}
	SetServiceTimeout("AgntChat", 120, 180)
	defer SetServiceTimeout("AgntChat", 120, 0)
	if chat.timeout() != 180 {
		t.Fatalf("override: got %d, want 180", chat.timeout())
	}
	if other.timeout() != Timeout || sameNameOtherArea.timeout() != Timeout {
		t.Fatalf("override leaked to another service: %d, %d", other.timeout(), sameNameOtherArea.timeout())
	}
	SetServiceTimeout("AgntChat", 120, 0)
	if chat.timeout() != Timeout {
		t.Fatalf("removed override: got %d, want the default %d", chat.timeout(), Timeout)
	}
}
