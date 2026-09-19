package network

import "testing"

func TestIsValidTarget(t *testing.T) {
	valid := []struct {
		name   string
		target string
	}{
		{"hostname", "example.com"},
		{"subdomain", "docs.devcockpit.app"},
		{"hyphenated host", "my-host.example.com"},
		{"ipv4", "192.168.1.1"},
		{"ipv4 loopback", "127.0.0.1"},
		{"ipv6 loopback", "::1"},
		{"ipv6 full", "2001:0db8:85a3:0000:0000:8a2e:0370:7334"},
		{"host with port", "example.com:8080"},
		{"single label", "localhost"},
		{"digits only", "12345"},
	}

	for _, tt := range valid {
		t.Run("valid/"+tt.name, func(t *testing.T) {
			if !isValidTarget(tt.target) {
				t.Errorf("isValidTarget(%q) = false, want true", tt.target)
			}
		})
	}

	invalid := []struct {
		name   string
		target string
	}{
		{"empty", ""},
		// Regression: the pattern was `[a-zA-Z0-9.-:]`, where `.-:` is a range
		// from '.' to ':' and silently admitted '/'.
		{"slash", "example.com/path"},
		{"bare slash", "/"},
		{"path traversal", "../etc/passwd"},
		{"space", "example.com evil"},
		{"semicolon", "example.com;whoami"},
		{"pipe", "example.com|whoami"},
		{"ampersand", "example.com&whoami"},
		{"dollar", "example.com$(whoami)"},
		{"backtick", "example.com`whoami`"},
		{"newline", "example.com\nevil"},
		{"null byte", "example.com\x00"},
		{"at sign", "user@example.com"},
		{"scheme", "https://example.com"},
		{"underscore", "bad_host.example.com"},
	}

	for _, tt := range invalid {
		t.Run("invalid/"+tt.name, func(t *testing.T) {
			if isValidTarget(tt.target) {
				t.Errorf("isValidTarget(%q) = true, want false", tt.target)
			}
		})
	}
}
