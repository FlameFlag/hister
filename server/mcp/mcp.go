// Package mcp exposes Hister's index through a Model Context Protocol server
// so LLM clients (Claude Desktop, ChatGPT Desktop, etc.) can query local
// browsing history.
//
// The MCP server is mounted as an http.Handler on the existing webserver's
// /mcp route. Authentication is delegated to the webserver's existing
// middleware: callers inject the authenticated UserID via WithUserID before
// invoking ServeHTTP, and tool handlers read it back via toolScope.userID.
package mcp

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/asciimoo/hister/config"
	"github.com/asciimoo/hister/server/document"
	"github.com/asciimoo/hister/server/indexer"
	"github.com/asciimoo/hister/server/model"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Options configures a Handler.
type Options struct {
	Version      string
	EnableWrites bool
}

// Handler returns an http.Handler implementing the MCP streamable-HTTP
// transport. It must be served behind the webserver's auth middleware:
// each request must carry an authenticated UserID on its context (via
// WithUserID) so multi-user calls can scope to the right account.
func Handler(cfg *config.Config, opts Options) http.Handler {
	scope := &toolScope{cfg: cfg, multiUser: cfg.App.UserHandling}
	srv := build(scope, opts)
	return mcpsdk.NewStreamableHTTPHandler(func(*http.Request) *mcpsdk.Server {
		return srv
	}, &mcpsdk.StreamableHTTPOptions{
		// Reclaims sessions from LLM clients that went quiet without
		// disconnecting. Distinct from ServerOptions.KeepAlive: KeepAlive
		// detects dead peers via ping; SessionTimeout closes peers that
		// are alive but idle.
		SessionTimeout: 10 * time.Minute,
		// Buffers outgoing SSE events per session so a client that
		// reconnects with Last-Event-ID can resume mid-stream. Without a
		// store, a dropped connection during a long index_url drops the
		// in-flight progress/log notifications.
		EventStore: mcpsdk.NewMemoryEventStore(nil),
	})
}

// keepAliveInterval pings idle clients and closes sessions that stop
// responding. A disconnected client would otherwise keep a session object
// alive indefinitely.
const keepAliveInterval = 30 * time.Second

func build(scope *toolScope, opts Options) *mcpsdk.Server {
	srv := mcpsdk.NewServer(
		&mcpsdk.Implementation{
			Name:       "hister",
			Version:    opts.Version,
			WebsiteURL: "https://github.com/asciimoo/hister",
		},
		&mcpsdk.ServerOptions{
			Instructions: serverInstructions,
			KeepAlive:    keepAliveInterval,
			// Bridges SDK-internal warnings (dropped notifications,
			// unhandled method errors) into our zerolog stream instead
			// of the stdlib default logger. The "source" tag lets us
			// grep SDK-origin lines out of our own log output.
			Logger:                  slog.New(zerolog.NewSlogHandler(log.Logger.With().Str("source", "mcp-sdk").Logger())),
			CompletionHandler:       completionHandler(scope),
			InitializedHandler:      initializedHandler(scope),
			RootsListChangedHandler: rootsListChangedHandler,
		},
	)

	srv.AddReceivingMiddleware(loggingMiddleware)
	srv.AddSendingMiddleware(sendingMiddleware)

	registerSearchTool(srv, scope)
	registerFetchDocumentTool(srv, scope)
	registerDocResourceTemplate(srv, scope)
	registerPrompts(srv)
	return srv
}

// userIDCtxKey types the context key used to bridge the authenticated
// UserID from the webserver into MCP tool handlers.
type userIDCtxKey struct{}

// WithUserID returns a context carrying uid for downstream MCP tool
// handlers to read. Call sites are the webserver's MCP route handler,
// after its own auth middleware has populated the user.
func WithUserID(ctx context.Context, uid uint) context.Context {
	return context.WithValue(ctx, userIDCtxKey{}, uid)
}

func userIDFromContext(ctx context.Context) uint {
	if v, ok := ctx.Value(userIDCtxKey{}).(uint); ok {
		return v
	}
	return 0
}

// toolScope carries per-server config plus per-request user resolution into
// each tool handler. It centralises the "whose data am I touching?" question
// so each tool doesn't re-derive it.
type toolScope struct {
	cfg       *config.Config
	multiUser bool
}

// userID resolves the effective UserID for a tool call. In single-user mode
// the returned UserID is 0 (no scoping). In multi-user mode it reads the
// UserID injected via WithUserID by the webserver's auth middleware.
// Returns (0, false) if multi-user mode has no identity attached, callers
// must refuse the request.
func (s *toolScope) userID(ctx context.Context) (uint, bool) {
	if !s.multiUser {
		return 0, true
	}
	if uid := userIDFromContext(ctx); uid > 0 {
		return uid, true
	}
	return 0, false
}

// rules returns the effective rule set for a UserID, mirroring the HTTP
// server's effectiveRules(): per-user rules if set, else the global ones.
func (s *toolScope) rules(uid uint) *config.Rules {
	if !s.multiUser || uid == 0 {
		return s.cfg.Rules
	}
	if r, err := model.GetUserRules(uid); err == nil && r != nil {
		return r
	}
	return s.cfg.Rules
}

// lookupDoc is the indexer's per-URL lookup scoped to the caller. In
// multi-user mode it pins to uid so another owner's copy of the same URL
// can't mask the caller's own document (doc IDs are "uid:url" but the
// url field is shared). In single-user mode the convention is uid=0,
// matching how indexer.GetByURLAndUser is called elsewhere in the tree
// (e.g. files.go).
func (s *toolScope) lookupDoc(url string, uid uint) *document.Document {
	if s.multiUser && uid > 0 {
		return indexer.GetByURLAndUser(url, uid)
	}
	return indexer.GetByURLAndUser(url, 0)
}
