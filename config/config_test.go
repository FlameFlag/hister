package config

import "testing"

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
