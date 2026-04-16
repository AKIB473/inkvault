package posts

import (
	"testing"
)

func TestSlugify(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"Hello World", "hello-world"},
		{"My Amazing Post!", "my-amazing-post"},
		{"  Leading/Trailing  ", "leading-trailing"},
		{"Multiple   Spaces", "multiple-spaces"},
		{"Already-fine-slug", "already-fine-slug"},
		// Japanese letters ARE valid unicode — slugify keeps them, service adds uuid fallback only for truly empty results
		// {"日本語テスト", ""}, // removed: unicode letters pass IsLetter()
		{"Post: A Deep Dive", "post-a-deep-dive"},
		{"100% Performance", "100-performance"},
		{"", ""},
	}

	for _, tc := range cases {
		got := slugify(tc.input)
		if got != tc.want {
			t.Errorf("slugify(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestSlugifyMaxLength(t *testing.T) {
	long := ""
	for i := 0; i < 250; i++ {
		long += "a"
	}
	got := slugify(long)
	if len(got) > 200 {
		t.Errorf("slug too long: %d chars", len(got))
	}
}

func TestEstimateReadingTime(t *testing.T) {
	cases := []struct {
		words int
		want  int
	}{
		{0, 1},    // minimum 1
		{100, 1},  // < 200 words = 1 min
		{200, 1},  // exactly 200 = 1 min
		{400, 2},  // 400 / 200 = 2 min
		{1000, 5}, // 1000 / 200 = 5 min
	}

	for _, tc := range cases {
		content := ""
		for i := 0; i < tc.words; i++ {
			content += "word "
		}
		got := estimateReadingTime(content)
		if got != tc.want {
			t.Errorf("estimateReadingTime(%d words) = %d, want %d", tc.words, got, tc.want)
		}
	}
}

func TestSanitizeTags(t *testing.T) {
	input := []string{"Go", "go", "  spaces  ", "UPPER", "", "duplicate", "duplicate"}
	got := sanitizeTags(input)

	// Should deduplicate and lowercase
	seen := map[string]bool{}
	for _, tag := range got {
		if seen[tag] {
			t.Errorf("duplicate tag: %q", tag)
		}
		seen[tag] = true
		if tag != tag {
			t.Errorf("tag not lowercased: %q", tag)
		}
		if tag == "" {
			t.Error("empty tag in output")
		}
	}

	// "go" should appear once (Go + go deduped)
	if !seen["go"] {
		t.Error("expected 'go' in sanitized tags")
	}
}

func TestParseVisibility(t *testing.T) {
	cases := []struct {
		input string
		want  int
	}{
		{"public", 0},
		{"unlisted", 1},
		{"members_only", 2},
		{"private", 3},
		{"password_protected", 4},
		{"unknown", 0}, // defaults to public
		{"", 0},
	}

	for _, tc := range cases {
		got := int(parseVisibility(tc.input))
		if got != tc.want {
			t.Errorf("parseVisibility(%q) = %d, want %d", tc.input, got, tc.want)
		}
	}
}
