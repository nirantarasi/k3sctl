package shellutil

import "testing"

func TestQuote(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "simple", input: "hello", want: "'hello'"},
		{name: "empty", input: "", want: "''"},
		{name: "with spaces", input: "hello world", want: "'hello world'"},
		{name: "with single quote", input: "it's", want: "'it'\\''s'"},
		{name: "path", input: "/var/lib/k3s", want: "'/var/lib/k3s'"},
		{name: "multiple quotes", input: "a'b'c", want: "'a'\\''b'\\''c'"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Quote(tt.input)
			if got != tt.want {
				t.Errorf("Quote(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
