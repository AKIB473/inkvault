package rss

import (
	"strings"
	"testing"
)

func TestXMLEscape(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"Hello & World", "Hello &amp; World"},
		{"<script>", "&lt;script&gt;"},
		{`say "hi"`, "say &quot;hi&quot;"},
		{"it's fine", "it&#39;s fine"},
		{"plain text", "plain text"},
	}
	for _, tc := range cases {
		got := xmlEscape(tc.input)
		if got != tc.want {
			t.Errorf("xmlEscape(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestXMLEscapeNoDoubleEncode(t *testing.T) {
	// Ensure we don't double-encode
	input := "a & b"
	got := xmlEscape(input)
	if strings.Contains(got, "&amp;amp;") {
		t.Errorf("double-encoded: %q", got)
	}
}
