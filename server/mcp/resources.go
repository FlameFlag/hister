package mcp

import (
	"context"
	"strings"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// docResourceTemplate is the URI template for indexed documents. `{+url}`
// uses RFC 6570 reserved expansion so the embedded URL keeps its ':' and
// '/' intact, a ref of `hister://doc/https://example.com/page` resolves
// by stripping docResourcePrefix from the URI.
const (
	docResourceTemplate = "hister://doc/{+url}"
	docResourcePrefix   = "hister://doc/"
)

func registerDocResourceTemplate(srv *mcpsdk.Server, scope *toolScope) {
	srv.AddResourceTemplate(&mcpsdk.ResourceTemplate{
		Name:        "document",
		Title:       "Indexed document",
		Description: "Extracted plain text of an indexed page. Reference as hister://doc/<full-url>.",
		URITemplate: docResourceTemplate,
		MIMEType:    "text/plain; charset=utf-8",
	}, func(ctx context.Context, req *mcpsdk.ReadResourceRequest) (*mcpsdk.ReadResourceResult, error) {
		uri := req.Params.URI
		docURL, ok := strings.CutPrefix(uri, docResourcePrefix)
		if !ok || docURL == "" {
			return nil, mcpsdk.ResourceNotFoundError(uri)
		}
		uid, ok := scope.userID(ctx)
		if !ok {
			return nil, errUnauthenticated
		}
		d := scope.lookupDoc(docURL, uid)
		if d == nil {
			return nil, mcpsdk.ResourceNotFoundError(uri)
		}
		return &mcpsdk.ReadResourceResult{
			Contents: []*mcpsdk.ResourceContents{{
				URI:      uri,
				MIMEType: "text/plain; charset=utf-8",
				Text:     d.Text,
			}},
		}, nil
	})
}
