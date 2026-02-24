package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// VerifyHMACSHA256 checks payload against a "sha256=<hex>" signature
// using the given secret. Returns false if the prefix is missing,
// the hex is invalid, or the HMAC does not match.
func VerifyHMACSHA256(payload []byte, signature, secret string) bool {
	if !strings.HasPrefix(signature, "sha256=") {
		return false
	}
	sig, err := hex.DecodeString(signature[7:])
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expected := mac.Sum(nil)
	return hmac.Equal(sig, expected)
}
