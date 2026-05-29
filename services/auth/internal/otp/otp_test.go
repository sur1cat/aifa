package otp

import (
	"testing"
)

func TestGenerateCode(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		code, err := generateCode()
		if err != nil {
			t.Fatalf("generateCode: %v", err)
		}
		if len(code) != 6 {
			t.Fatalf("expected 6 chars, got %d: %q", len(code), code)
		}
		for _, ch := range code {
			if ch < '0' || ch > '9' {
				t.Fatalf("non-digit in code: %q", code)
			}
		}
		seen[code] = true
	}
	// With 100 random 6-digit codes, collisions are possible but
	// we should see at least 80 distinct values.
	if len(seen) < 80 {
		t.Fatalf("expected at least 80 unique codes from 100 generations, got %d", len(seen))
	}
}
