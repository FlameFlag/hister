package mcp

import (
	"strings"
	"testing"

	"github.com/asciimoo/hister/server/document"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func Test_searchContent_empty(t *testing.T) {
	// No hits -> nil Content so the SDK falls back to the serialized
	// structured-output TextContent block.
	if c := searchContent(nil); c != nil {
		t.Errorf("searchContent(nil) = %v, want nil", c)
	}
	if c := searchContent([]SearchHit{}); c != nil {
		t.Errorf("searchContent([]) = %v, want nil", c)
	}
}

func Test_searchContent_buildsResourceLinks(t *testing.T) {
	hits := []SearchHit{
		{URL: "https://example.com/a", Title: "Article A", Snippet: "snippet A"},
		{URL: "https://example.com/b", Title: "", Snippet: "snippet B"}, // fallback Name=URL
	}
	content := searchContent(hits)
	if len(content) != len(hits) {
		t.Fatalf("len(content) = %d, want %d", len(content), len(hits))
	}
	for i, c := range content {
		link, ok := c.(*mcpsdk.ResourceLink)
		if !ok {
			t.Fatalf("content[%d] is %T, want *ResourceLink", i, c)
		}
		wantURI := docResourcePrefix + hits[i].URL
		if link.URI != wantURI {
			t.Errorf("content[%d].URI = %q, want %q", i, link.URI, wantURI)
		}
		if link.Description != hits[i].Snippet {
			t.Errorf("content[%d].Description = %q, want %q", i, link.Description, hits[i].Snippet)
		}
		if link.MIMEType != "text/plain; charset=utf-8" {
			t.Errorf("content[%d].MIMEType = %q", i, link.MIMEType)
		}
	}
	// Name falls back to URL when title is empty.
	if name := content[1].(*mcpsdk.ResourceLink).Name; name != hits[1].URL {
		t.Errorf("content[1].Name = %q, want URL fallback %q", name, hits[1].URL)
	}
	// Title is left empty (not URL-substituted) so clients can distinguish
	// untitled docs from titled ones if they want to.
	if title := content[1].(*mcpsdk.ResourceLink).Title; title != "" {
		t.Errorf("content[1].Title = %q, want empty", title)
	}
}

func Test_toHit_snippetTruncation(t *testing.T) {
	short := strings.Repeat("a", 10)
	h := toHit(&document.Document{URL: "u", Text: short})
	if h.Snippet != short || h.SnippetTruncated {
		t.Errorf("short: Snippet=%q Truncated=%v, want pass-through", h.Snippet, h.SnippetTruncated)
	}

	long := strings.Repeat("a", snippetMaxChars+50)
	h = toHit(&document.Document{URL: "u", Text: long})
	if !h.SnippetTruncated {
		t.Errorf("long: Truncated=false, want true")
	}
	if n := len([]rune(h.Snippet)); n != snippetMaxChars {
		t.Errorf("long: rune count = %d, want %d", n, snippetMaxChars)
	}

	// Multi-byte runes: the truncation must count runes, not bytes, so the
	// snippet stays valid UTF-8 and doesn't split a codepoint.
	multibyte := strings.Repeat("日", snippetMaxChars+10)
	h = toHit(&document.Document{URL: "u", Text: multibyte})
	if !h.SnippetTruncated {
		t.Errorf("multibyte: Truncated=false, want true")
	}
	if n := len([]rune(h.Snippet)); n != snippetMaxChars {
		t.Errorf("multibyte: rune count = %d, want %d", n, snippetMaxChars)
	}
	// Validate no partial codepoints: a round-trip through []rune is lossy
	// on invalid UTF-8 but identity-preserving on valid input.
	if string([]rune(h.Snippet)) != h.Snippet {
		t.Errorf("multibyte: snippet is not valid UTF-8 after truncation")
	}
}

func Test_newMCPIndexQuery_starUsesMatchAllWithoutHighlight(t *testing.T) {
	q := newMCPIndexQuery(SearchArgs{
		Text:     " * ",
		Sort:     "domain",
		Limit:    50,
		PageKey:  `["example.com","https://example.com"]`,
		DateFrom: 100,
		DateTo:   200,
	}, 42, 50)

	if !q.MatchAll {
		t.Fatalf("MatchAll = false, want true")
	}
	if q.Highlight != "" {
		t.Fatalf("Highlight = %q, want empty", q.Highlight)
	}
	if q.Text != " * " || q.Sort != "domain" || q.Limit != 50 || q.UserID != 42 || !q.Facets {
		t.Fatalf("query did not preserve search args: %+v", q)
	}
	if q.DateFrom != 100 || q.DateTo != 200 {
		t.Fatalf("date bounds = %d..%d, want 100..200", q.DateFrom, q.DateTo)
	}
}

func Test_newMCPIndexQuery_normalSearchKeepsHighlight(t *testing.T) {
	q := newMCPIndexQuery(SearchArgs{Text: "golang"}, 42, 20)
	if q.MatchAll {
		t.Fatalf("MatchAll = true, want false")
	}
	if q.Highlight != "HTML" {
		t.Fatalf("Highlight = %q, want HTML", q.Highlight)
	}
}
