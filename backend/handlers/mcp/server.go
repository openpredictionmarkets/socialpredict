// Package mcp is the read-only MCP (Model Context Protocol) transport adapter.
//
// It exposes public market data to MCP-compatible AI assistant clients over
// the Streamable HTTP transport, mounted by the HTTP server at /v0/mcp. The
// adapter consumes the markets domain service through a narrow, consumer-defined
// interface and registers only read-only tools: no capability in this package
// may mutate platform state or expose data invisible to a logged-out visitor.
//
// See specs/001-mcp-server/ for the feature specification and contracts.
package mcp
