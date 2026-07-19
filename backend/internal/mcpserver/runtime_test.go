package mcpserver

import (
	"context"
	"reflect"
	"sort"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestRegisterToolsExposesPublicMarketTools(t *testing.T) {
	ctx := context.Background()
	server := NewRuntime(&marketToolService{}, nil).MCPServer()
	client := mcp.NewClient(&mcp.Implementation{Name: "registration-test", Version: "v0.1.0"}, nil)
	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	serverSession, err := server.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("server.Connect returned error: %v", err)
	}
	defer serverSession.Close()
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client.Connect returned error: %v", err)
	}
	defer clientSession.Close()

	result, err := clientSession.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("ListTools returned error: %v", err)
	}
	got := make([]string, 0, len(result.Tools))
	for _, tool := range result.Tools {
		got = append(got, tool.Name)
	}
	sort.Strings(got)
	want := []string{
		"get_market",
		"get_market_discovery",
		"get_market_summary",
		"get_market_tag",
		"list_market_tags",
		"list_markets",
		"quote_market_probability",
		"search_markets",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("registered tool names = %v, want %v", got, want)
	}
}
