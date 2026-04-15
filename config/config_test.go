package config

import (
	"regexp"
	"testing"
)

func TestBasePathPrefix(t *testing.T) {
	tests := []struct {
		name   string
		base   string
		prefix string
	}{
		{name: "root-no-slash", base: "https://example.com", prefix: ""},
		{name: "root-with-slash", base: "https://example.com/", prefix: ""},
		{name: "subfolder", base: "https://example.com/subfolder", prefix: "/subfolder"},
		{name: "subfolder-trailing", base: "https://example.com/subfolder/", prefix: "/subfolder"},
		{name: "nested", base: "https://example.com/a/b", prefix: "/a/b"},
		{name: "nested-trailing", base: "https://example.com/a/b/", prefix: "/a/b"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Server: Server{BaseURL: tt.base}}
			if got := cfg.BasePathPrefix(); got != tt.prefix {
				t.Fatalf("BasePathPrefix()=%q, want %q", got, tt.prefix)
			}
		})
	}
}

func TestSensitiveContentPatterns(t *testing.T) {
	patterns := CreateDefaultConfig().SensitiveContentPatterns
	tests := []struct {
		name    string
		pattern string
		input   string
		match   bool
	}{
		{name: "aws_access_key/quoted", pattern: "aws_access_key", input: `key: "AKIAIOSFODNN7EXAMPLE"`, match: true},
		{name: "aws_access_key/whitespace", pattern: "aws_access_key", input: "token AKIAIOSFODNN7EXAMPLE end", match: true},
		{name: "aws_access_key/single-quoted", pattern: "aws_access_key", input: `'AKIAIOSFODNN7EXAMPLE'`, match: true},
		{name: "aws_access_key/start-of-string", pattern: "aws_access_key", input: "AKIAIOSFODNN7EXAMPLE ", match: true},
		{name: "aws_access_key/end-of-string", pattern: "aws_access_key", input: " AKIAIOSFODNN7EXAMPLE", match: true},
		{name: "aws_access_key/base64-blob", pattern: "aws_access_key", input: "d09GMgABAAAAAKIAIOSFODNN7EXAMPLEXYZABCDEF", match: false},
		{name: "aws_access_key/css-font", pattern: "aws_access_key", input: "url(data:font/woff2;base64,d09GMgABAAAAAKIA1234567890ABCDEF)", match: false},
		{name: "github_token/valid", pattern: "github_token", input: "ghp_abcdefghijklmnopqrstuvwxyzABCDEFGHIJ", match: true},
		{name: "generic_private_key", pattern: "generic_private_key", input: "-----BEGIN RSA PRIVATE KEY-----", match: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw, ok := patterns[tt.pattern]
			if !ok {
				t.Fatalf("pattern %q not in defaults", tt.pattern)
			}
			re := regexp.MustCompile(raw)
			if got := re.MatchString(tt.input); got != tt.match {
				t.Fatalf("MatchString(%q) = %v, want %v", tt.input, got, tt.match)
			}
		})
	}
}

func TestWebSocketURLHonorsBasePath(t *testing.T) {
	tests := []struct {
		name string
		base string
		want string
	}{
		{name: "http-root", base: "http://example.com:1234", want: "ws://example.com:1234/search"},
		{name: "https-root", base: "https://example.com", want: "wss://example.com/search"},
		{name: "http-subfolder", base: "http://example.com/subfolder", want: "ws://example.com/subfolder/search"},
		{name: "https-nested", base: "https://example.com/a/b/", want: "wss://example.com/a/b/search"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Server: Server{BaseURL: tt.base}}
			if got := cfg.WebSocketURL(); got != tt.want {
				t.Fatalf("WebSocketURL()=%q, want %q", got, tt.want)
			}
		})
	}
}
