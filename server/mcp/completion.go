package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/asciimoo/hister/server/indexer"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// maxCompletionValues caps how many suggestions we return per call. The spec
// allows up to 100; clients usually paginate by typing more characters, so
// a smaller cap keeps responses snappy and avoids dominating context.
const maxCompletionValues = 20

// completionSource resolves values for a single (ref, argument) pair. The
// dispatch table below is the single source of truth for what's
// completable: adding a new prompt argument means adding one row, not
// extending a switch.
type completionSource func(scope *toolScope, uid uint, value string) (*mcpsdk.CompleteResult, error)

// completionKey identifies an argument eligible for completion. For
// ref/prompt the id field is the prompt Name; for ref/resource it's the
// resource-template URI.
type completionKey struct {
	refType, id, argName string
}

var completionSources = map[completionKey]completionSource{
	{"ref/prompt", promptWhatDoIKnow, "topic"}:     topicCompletions,
	{"ref/prompt", promptSummarizeRecent, "topic"}: topicCompletions,
	{"ref/resource", docResourceTemplate, "url"}:   urlCompletions,
}

// completionHandler returns autocomplete suggestions for prompt arguments
// and the doc resource template's {url} variable. Values are pulled from
// the index's own facets and URL keyspace, so completions are grounded in
// what the user actually has indexed rather than generic lists.
func completionHandler(scope *toolScope) func(context.Context, *mcpsdk.CompleteRequest) (*mcpsdk.CompleteResult, error) {
	return func(ctx context.Context, req *mcpsdk.CompleteRequest) (*mcpsdk.CompleteResult, error) {
		ref := req.Params.Ref
		if ref == nil {
			return emptyCompletion(), nil
		}
		uid, ok := scope.userID(ctx)
		if !ok {
			return nil, errUnauthenticated
		}
		id := ref.Name
		if ref.Type == "ref/resource" {
			id = ref.URI
		}
		fn, ok := completionSources[completionKey{ref.Type, id, req.Params.Argument.Name}]
		if !ok {
			return emptyCompletion(), nil
		}
		return fn(scope, uid, strings.TrimSpace(req.Params.Argument.Value))
	}
}

// topicCompletions surfaces indexed domains whose name starts with the
// user's partial input. Domains are the cheapest useful signal ("what
// topic?" roughly maps to "which site?") and are already in the facets.
//
// Asks the indexer for a larger-than-default facet pool so a user typing
// a prefix that isn't in the global top-10 domains still gets completions.
// We over-fetch and post-filter; the bleve top-N is counted per-shard so
// a slightly generous cap is cheap.
func topicCompletions(scope *toolScope, uid uint, prefix string) (*mcpsdk.CompleteResult, error) {
	res, err := indexer.Search(scope.cfg, &indexer.Query{
		MatchAll:      true,
		UserID:        uid,
		Facets:        true,
		FacetTermSize: 200,
		Limit:         1,
	})
	if err != nil {
		return nil, fmt.Errorf("completion: %w", err)
	}
	values := make([]string, 0, maxCompletionValues)
	if res.Facets != nil {
		for _, tc := range res.Facets.Domains {
			if prefix != "" && !strings.HasPrefix(tc.Term, prefix) {
				continue
			}
			values = append(values, tc.Term)
			if len(values) >= maxCompletionValues {
				break
			}
		}
	}
	return &mcpsdk.CompleteResult{
		Completion: mcpsdk.CompletionResultDetails{
			Values: values,
			Total:  len(values),
		},
	}, nil
}

// urlCompletions returns indexed URLs matching the partial input. Uses the
// search DSL's url:<prefix>* so results are bounded by the URL field and
// not by full-text scoring. Prefixes that contain DSL metacharacters are
// rejected (empty result) rather than interpolated; otherwise a space,
// quote, paren, or pipe in the user's input would re-split the query and
// produce surprising completions.
func urlCompletions(scope *toolScope, uid uint, prefix string) (*mcpsdk.CompleteResult, error) {
	q := &indexer.Query{
		UserID: uid,
		Limit:  maxCompletionValues,
	}
	if prefix == "" {
		q.MatchAll = true
	} else {
		if !isSafeURLPrefix(prefix) {
			return emptyCompletion(), nil
		}
		q.Text = "url:" + prefix + "*"
	}
	res, err := indexer.Search(scope.cfg, q)
	if err != nil {
		return nil, fmt.Errorf("completion: %w", err)
	}
	values := make([]string, 0, len(res.Documents))
	for _, d := range res.Documents {
		values = append(values, d.URL)
	}
	return &mcpsdk.CompleteResult{
		Completion: mcpsdk.CompletionResultDetails{
			Values: values,
			Total:  len(values),
		},
	}, nil
}

// isSafeURLPrefix returns true when the prefix contains only characters
// that are both valid in a URL and safe to interpolate into the DSL token
// `url:<prefix>*` without being reinterpreted. DSL-special characters
// (whitespace, quotes, parens, pipes, a leading dash) would otherwise
// escape the url: field qualifier and change the query's meaning.
func isSafeURLPrefix(s string) bool {
	if s == "" || s[0] == '-' {
		return false
	}
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z',
			r >= 'A' && r <= 'Z',
			r >= '0' && r <= '9':
		case strings.ContainsRune(":/._-~%?#&=@+", r):
		default:
			return false
		}
	}
	return true
}

func emptyCompletion() *mcpsdk.CompleteResult {
	return &mcpsdk.CompleteResult{
		Completion: mcpsdk.CompletionResultDetails{Values: []string{}},
	}
}
