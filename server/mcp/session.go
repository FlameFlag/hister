package mcp

import (
	"context"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rs/zerolog/log"
)

// clientCapabilities returns the peer's advertised capability set, or nil
// if the session hasn't finished initialize or the client offered none.
// Call sites gate optional features (sampling, elicitation, …) on a
// non-nil sub-field; never send them unless the client opted in.
func clientCapabilities(ip *mcpsdk.InitializeParams) *mcpsdk.ClientCapabilities {
	if ip == nil {
		return nil
	}
	return ip.Capabilities
}

// initializedHandler runs once per session after the MCP handshake. It logs
// who connected and which optional capabilities they advertised (sampling,
// elicitation, roots), useful when debugging why a feature path didn't
// fire for a given client. If the client exposes roots, we ask for them
// here so the initial list is visible without waiting for a change
// notification.
func initializedHandler(_ *toolScope) func(context.Context, *mcpsdk.InitializedRequest) {
	return func(ctx context.Context, req *mcpsdk.InitializedRequest) {
		if req == nil || req.Session == nil {
			return
		}
		ip := req.Session.InitializeParams()
		ev := log.Info().Str("session", req.Session.ID())
		if ip != nil && ip.ClientInfo != nil {
			ev = ev.
				Str("client_name", ip.ClientInfo.Name).
				Str("client_version", ip.ClientInfo.Version)
		}
		if caps := clientCapabilities(ip); caps != nil {
			ev = ev.
				Bool("sampling", caps.Sampling != nil).
				Bool("elicitation", caps.Elicitation != nil).
				Bool("roots", caps.RootsV2 != nil)
		}
		ev.Msg("mcp session initialized")

		// Detach: ListRoots is a round-trip back to the client and must not
		// block the initialize callback.
		go logRoots(context.WithoutCancel(ctx), req.Session, "initial")
	}
}

// rootsListChangedHandler re-fetches the client's roots when it signals a
// change. We don't act on roots today (indexed URIs aren't file://), but
// logging them makes it easier to reason about future workspace-scoping
// features and surfaces any client that advertises listChanged but sends
// an empty list.
func rootsListChangedHandler(ctx context.Context, req *mcpsdk.RootsListChangedRequest) {
	if req == nil || req.Session == nil {
		return
	}
	logRoots(ctx, req.Session, "changed")
}

// logRoots calls ListRoots and writes a single log line. No-op if the
// client didn't advertise the roots capability, calling ListRoots on a
// client that doesn't support it returns an error we can safely ignore
// rather than noisily log.
func logRoots(ctx context.Context, ss *mcpsdk.ServerSession, reason string) {
	caps := clientCapabilities(ss.InitializeParams())
	if caps == nil || caps.RootsV2 == nil {
		return
	}
	res, err := ss.ListRoots(ctx, nil)
	if err != nil {
		log.Debug().Err(err).Str("session", ss.ID()).Msg("mcp: list roots failed")
		return
	}
	uris := make([]string, 0, len(res.Roots))
	for _, r := range res.Roots {
		uris = append(uris, r.URI)
	}
	log.Info().
		Str("session", ss.ID()).
		Str("reason", reason).
		Strs("roots", uris).
		Msg("mcp client roots")
}
