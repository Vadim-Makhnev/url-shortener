package service

import (
	"strings"
	"testing"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func TestService_GenerateShortCode(t *testing.T) {
	code := generateShortCode()

	if len(code) != 6 {
		t.Errorf("length should be 6 symbols, got %d", len(code))
	}

	codes := make(map[string]bool)

	for i := 0; i < 100; i++ {
		code := generateShortCode()

		if len(code) != 6 {
			t.Errorf("iteration %d: length should be 6 symbols, got %d", i, len(code))
		}

		for _, char := range code {
			if !strings.Contains(charset, string(char)) {
				t.Errorf("invalid character %q in code %s", char, code)
			}
		}

		if _, ok := codes[code]; ok {
			t.Errorf("duplicate code generated: %s", code)
		}

		codes[code] = true
	}
}
