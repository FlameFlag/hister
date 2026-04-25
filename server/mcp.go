// SPDX-License-Identifier: AGPL-3.0-or-later

package server

import (
	"net/http"
	"sync"

	histermcp "github.com/asciimoo/hister/server/mcp"
)

// mcpHandler is the SDK-backed MCP streamable-HTTP handler. Built lazily on
// first request: Version and config are only fully populated by the time
// Listen() is dispatching to the routing tree, and a server that never
// receives an /mcp request shouldn't pay for it at startup.
var (
	mcpHandlerOnce sync.Once
	mcpHandler     http.Handler
)

func serveMCP(c *webContext) {
	mcpHandlerOnce.Do(func() {
		mcpHandler = histermcp.Handler(c.Config, histermcp.Options{
			Version:      Version,
			EnableWrites: c.Config.App.EnableMCPWrites,
		})
	})
	ctx := histermcp.WithUserID(c.Request.Context(), c.UserID)
	mcpHandler.ServeHTTP(c.Response, c.Request.WithContext(ctx))
}
