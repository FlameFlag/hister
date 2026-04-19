package mcp

import (
	"context"
	"time"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rs/zerolog/log"
)

// logMCPCall writes one zerolog line describing a completed JSON-RPC call.
// Shared between the receiving and sending middlewares so the two paths
// log identical fields (method, session, duration, error); label is the
// only thing that differs ("mcp request" vs "mcp send"). Tool-specific
// fields are populated when the request/result shapes match, a no-op
// otherwise, so the sending path (which never sees a CallToolRequest)
// pays nothing for the check.
func logMCPCall(req mcpsdk.Request, method string, result mcpsdk.Result, err error, start time.Time, label string) {
	ev := log.Debug()
	if err != nil {
		ev = log.Warn().Err(err)
	}
	ev = ev.Str("method", method).Dur("duration", time.Since(start))
	if sess := req.GetSession(); sess != nil {
		if id := sess.ID(); id != "" {
			ev = ev.Str("session", id)
		}
	}
	if call, ok := req.(*mcpsdk.CallToolRequest); ok && call.Params != nil {
		ev = ev.Str("tool", call.Params.Name)
	}
	if res, ok := result.(*mcpsdk.CallToolResult); ok && res != nil && res.IsError {
		ev = ev.Bool("tool_error", true)
	}
	ev.Msg(label)
}

// loggingMiddleware logs every incoming JSON-RPC method call. Tool-call
// errors raised via CallToolResult.IsError (rather than a Go error) are
// still surfaced via the tool_error field.
//
// The session ID is only populated on the streamable-HTTP transport;
// stdio sessions ID is empty, which is fine to log as such.
func loggingMiddleware(next mcpsdk.MethodHandler) mcpsdk.MethodHandler {
	return func(ctx context.Context, method string, req mcpsdk.Request) (mcpsdk.Result, error) {
		start := time.Now()
		result, err := next(ctx, method, req)
		logMCPCall(req, method, result, err, start, "mcp request")
		return result, err
	}
}

// sendingMiddleware logs every outgoing server->client call (notifications,
// ListRoots, CreateMessage, Elicit, …). Symmetric to loggingMiddleware for
// the inbound path and useful for explaining transport-side hiccups,
// e.g. a ResourceUpdated notification that's slow because the client
// stalled on SSE flush. Keep-alive pings are dropped to avoid flooding
// the log at the 30s interval.
func sendingMiddleware(next mcpsdk.MethodHandler) mcpsdk.MethodHandler {
	return func(ctx context.Context, method string, req mcpsdk.Request) (mcpsdk.Result, error) {
		if method == "ping" {
			return next(ctx, method, req)
		}
		start := time.Now()
		result, err := next(ctx, method, req)
		logMCPCall(req, method, result, err, start, "mcp send")
		return result, err
	}
}
