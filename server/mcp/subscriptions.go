package mcp

import (
	"context"
	"fmt"
	"strings"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rs/zerolog/log"
)

// subscribeHandler authorises a resources/subscribe request. The SDK keeps
// the per-session subscription set itself; this handler only validates
// the target: the URI must match our doc resource template, and in
// multi-user mode the caller must own a document at that URL. Without
// that ownership check a user could subscribe to an arbitrary URL and
// learn (via the timing of updates) whether another user indexed it.
func subscribeHandler(scope *toolScope) func(context.Context, *mcpsdk.SubscribeRequest) error {
	return func(ctx context.Context, req *mcpsdk.SubscribeRequest) error {
		uri := req.Params.URI
		docURL, ok := strings.CutPrefix(uri, docResourcePrefix)
		if !ok {
			return fmt.Errorf("only %s* URIs are subscribable", docResourcePrefix)
		}
		if docURL == "" {
			return fmt.Errorf("empty document URL in %q", uri)
		}
		uid, ok := scope.userID(ctx)
		if !ok {
			return errUnauthenticated
		}
		d := scope.lookupDoc(docURL, uid)
		if d == nil {
			// Same message as ReadResource's not-found so callers can't
			// use subscribe to probe for documents they can't read.
			return mcpsdk.ResourceNotFoundError(uri)
		}
		return nil
	}
}

// unsubscribeHandler accepts all unsubscribe requests, the SDK removes
// the entry from its internal map regardless of what we return, and
// there's no per-session state we keep outside of it.
func unsubscribeHandler(_ context.Context, _ *mcpsdk.UnsubscribeRequest) error {
	return nil
}

// notifyDocChanged sends a resources/updated notification for a
// document URL to every subscribed session. Safe to call with a nil
// server (stdio handlers built before the server is wired for tests).
// Non-fatal: notification failures are logged but don't propagate,
// since the underlying write already succeeded.
func notifyDocChanged(ctx context.Context, srv *mcpsdk.Server, docURL string) {
	if srv == nil || docURL == "" {
		return
	}
	uri := docResourcePrefix + docURL
	if err := srv.ResourceUpdated(ctx, &mcpsdk.ResourceUpdatedNotificationParams{URI: uri}); err != nil {
		log.Debug().Err(err).Str("uri", uri).Msg("mcp: resource updated notify failed")
	}
}
