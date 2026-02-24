package tests

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/saichler/l8web/go/web/webhook"
)

func computeHMAC(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func TestVerifyHMACSHA256_Valid(t *testing.T) {
	payload := []byte(`{"action":"opened"}`)
	secret := "test-secret"
	sig := computeHMAC(payload, secret)

	if !webhook.VerifyHMACSHA256(payload, sig, secret) {
		t.Fatal("expected valid signature to return true")
	}
}

func TestVerifyHMACSHA256_Invalid(t *testing.T) {
	payload := []byte(`{"action":"opened"}`)
	secret := "test-secret"
	sig := computeHMAC(payload, "wrong-secret")

	if webhook.VerifyHMACSHA256(payload, sig, secret) {
		t.Fatal("expected invalid signature to return false")
	}
}

func TestVerifyHMACSHA256_BadPrefix(t *testing.T) {
	payload := []byte(`{"action":"opened"}`)
	secret := "test-secret"

	if webhook.VerifyHMACSHA256(payload, "md5=abcdef", secret) {
		t.Fatal("expected missing sha256= prefix to return false")
	}
}

func TestVerifyHMACSHA256_BadHex(t *testing.T) {
	payload := []byte(`{"action":"opened"}`)
	secret := "test-secret"

	if webhook.VerifyHMACSHA256(payload, "sha256=not-valid-hex!!", secret) {
		t.Fatal("expected invalid hex to return false")
	}
}
