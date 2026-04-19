package mcp

import (
	"cmp"
	"context"
	"errors"
	"strings"
	"time"

	"github.com/asciimoo/hister/server/document"
	"github.com/asciimoo/hister/server/indexer"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	snippetMaxChars = 400
	defaultPageSize = 20
	maxPageSize     = 100
)

// errUnauthenticated is returned by tool handlers when multi-user mode is on
// but no authenticated UserID reached the handler (e.g. stdio without --user,
// or an internal misconfiguration, bearerAuth should have already rejected
// unauthenticated HTTP callers before they get this far).
var errUnauthenticated = errors.New("no authenticated user on request; multi-user MCP requires per-user bearer tokens (HTTP) or --user (stdio)")

type SearchArgs struct {
	Text     string `json:"text" jsonschema:"query in Hister's DSL (see tool description for examples)"`
	Sort     string `json:"sort,omitempty" jsonschema:"sort order: score (default) or domain"`
	Limit    int    `json:"limit,omitempty" jsonschema:"maximum hits to return (default 20, max 100)"`
	PageKey  string `json:"page_key,omitempty" jsonschema:"cursor from a previous response; pass back verbatim to paginate"`
	DateFrom int64  `json:"date_from,omitempty" jsonschema:"restrict to documents added at or after this unix timestamp"`
	DateTo   int64  `json:"date_to,omitempty" jsonschema:"restrict to documents added at or before this unix timestamp"`
}

type SearchHit struct {
	URL              string  `json:"url"`
	Title            string  `json:"title,omitempty"`
	Domain           string  `json:"domain,omitempty"`
	Score            float64 `json:"score,omitempty"`
	Added            int64   `json:"added,omitempty"`
	Language         string  `json:"language,omitempty"`
	Snippet          string  `json:"snippet,omitempty"`
	SnippetTruncated bool    `json:"snippet_truncated,omitempty"`
	// textBytes is the full plain-text size of the indexed document,
	// used to populate ResourceLink.Size so clients can budget context
	// before fetching the document. Not serialised to structured output
	// (snippet already gives the model a size sense).
	textBytes int64
}

type SearchResult struct {
	Total   uint64                `json:"total"`
	Hits    []SearchHit           `json:"hits"`
	Facets  *indexer.FacetsResult `json:"facets,omitempty"`
	PageKey string                `json:"page_key,omitempty"`
	Hint    string                `json:"hint,omitempty"`
}

func registerSearchTool(srv *mcpsdk.Server, scope *toolScope) {
	mcpsdk.AddTool(srv, &mcpsdk.Tool{
		Name:        "search",
		Title:       "Search browsing history",
		Description: searchDesc,
		Annotations: &mcpsdk.ToolAnnotations{
			ReadOnlyHint:   true,
			IdempotentHint: true,
			OpenWorldHint:  new(false), // only the local index
		},
	}, func(ctx context.Context, _ *mcpsdk.CallToolRequest, args SearchArgs) (*mcpsdk.CallToolResult, SearchResult, error) {
		if strings.TrimSpace(args.Text) == "" {
			return nil, SearchResult{}, errors.New("text is required; pass at least one search term (e.g. \"domain:example.com\")")
		}
		uid, ok := scope.userID(ctx)
		if !ok {
			return nil, SearchResult{}, errUnauthenticated
		}
		limit := args.Limit
		if limit <= 0 {
			limit = defaultPageSize
		}
		limit = min(limit, maxPageSize)
		q := newMCPIndexQuery(args, uid, limit)
		res, err := indexer.Search(scope.cfg, q)
		if err != nil {
			return nil, SearchResult{}, err
		}
		out := SearchResult{Total: res.Total, PageKey: res.PageKey, Facets: res.Facets}
		for _, d := range res.Documents {
			out.Hits = append(out.Hits, toHit(d))
		}
		if res.Total == 0 {
			out.Hint = "no matches; try broader terms, drop a field qualifier, or check spelling"
		}
		// Emit a ResourceLink per hit alongside the structured output so
		// capable clients can surface each URL as a first-class resource
		// (openable via the hister://doc/<url> template) without parsing
		// structuredContent.
		return &mcpsdk.CallToolResult{Content: searchContent(out.Hits)}, out, nil
	})
}

func newMCPIndexQuery(args SearchArgs, uid uint, limit int) *indexer.Query {
	q := &indexer.Query{
		Text:     args.Text,
		Sort:     args.Sort,
		Limit:    limit,
		PageKey:  args.PageKey,
		DateFrom: args.DateFrom,
		DateTo:   args.DateTo,
		UserID:   uid,
		Facets:   true,
	}
	if strings.TrimSpace(args.Text) == "*" {
		q.MatchAll = true
	} else {
		q.Highlight = "HTML"
	}
	return q
}

// searchContent builds ResourceLink content entries for each hit. Each link
// resolves through the doc resource template, letting a client read full
// text without a second tool call. Returns nil when there are no hits so
// the SDK falls back to the serialized-structuredContent text block.
func searchContent(hits []SearchHit) []mcpsdk.Content {
	if len(hits) == 0 {
		return nil
	}
	content := make([]mcpsdk.Content, 0, len(hits))
	for _, h := range hits {
		link := &mcpsdk.ResourceLink{
			URI:         docResourcePrefix + h.URL,
			Name:        cmp.Or(h.Title, h.URL),
			Title:       h.Title,
			Description: h.Snippet,
			MIMEType:    "text/plain; charset=utf-8",
		}
		// Size is the full plain-text byte count so clients can budget
		// context window usage before reading the resource. Distinct
		// from the snippet, which is already truncated.
		if h.textBytes > 0 {
			size := h.textBytes
			link.Size = &size
		}
		if h.Added > 0 {
			link.Annotations = &mcpsdk.Annotations{
				LastModified: time.Unix(h.Added, 0).UTC().Format(time.RFC3339),
			}
		}
		content = append(content, link)
	}
	return content
}

func toHit(d *document.Document) SearchHit {
	snippet, truncated := truncateRunes(d.Text, snippetMaxChars)
	return SearchHit{
		URL:              d.URL,
		Title:            d.Title,
		Domain:           d.Domain,
		Score:            d.Score,
		Added:            d.Added,
		Language:         d.Language,
		textBytes:        int64(len(d.Text)),
		Snippet:          snippet,
		SnippetTruncated: truncated,
	}
}

// truncateRunes returns s cut to at most max runes and a flag indicating
// whether truncation happened. Splits on a rune boundary (range over a
// string yields byte offsets of rune starts) so the result is valid UTF-8.
func truncateRunes(s string, max int) (string, bool) {
	n := 0
	for i := range s {
		if n == max {
			return s[:i], true
		}
		n++
	}
	return s, false
}

// stripHTML returns a shallow copy of d with HTML cleared, so callers can
// return it over the wire without mutating the original (which may be a fresh
// document from GetByURL that we don't want to scribble on).
func stripHTML(d *document.Document) *document.Document {
	cp := *d
	cp.HTML = ""
	return &cp
}

type FetchDocumentArgs struct {
	URL         string `json:"url" jsonschema:"exact URL of the document to fetch"`
	IncludeHTML bool   `json:"include_html,omitempty" jsonschema:"include raw HTML in the response"`
}

type FetchDocumentResult struct {
	Found bool               `json:"found"`
	Doc   *document.Document `json:"doc,omitempty"`
	Hint  string             `json:"hint,omitempty"`
}

func registerFetchDocumentTool(srv *mcpsdk.Server, scope *toolScope) {
	mcpsdk.AddTool(srv, &mcpsdk.Tool{
		Name:        "fetch_document",
		Title:       "Fetch indexed document by URL",
		Description: fetchDocumentDesc,
		Annotations: &mcpsdk.ToolAnnotations{
			ReadOnlyHint:   true,
			IdempotentHint: true,
			OpenWorldHint:  new(false),
		},
	}, func(ctx context.Context, _ *mcpsdk.CallToolRequest, args FetchDocumentArgs) (*mcpsdk.CallToolResult, FetchDocumentResult, error) {
		if strings.TrimSpace(args.URL) == "" {
			return nil, FetchDocumentResult{}, errors.New("url is required")
		}
		uid, ok := scope.userID(ctx)
		if !ok {
			return nil, FetchDocumentResult{}, errUnauthenticated
		}
		d := scope.lookupDoc(args.URL, uid)
		if d == nil {
			return nil, FetchDocumentResult{Found: false, Hint: "URL not in index; use search to find related documents or index_url to add it"}, nil
		}
		if !args.IncludeHTML {
			d = stripHTML(d)
		}
		return nil, FetchDocumentResult{Found: true, Doc: d}, nil
	})
}
