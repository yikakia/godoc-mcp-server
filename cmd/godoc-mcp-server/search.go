package main

import (
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/pkg/errors"
	"github.com/yikakia/godoc-mcp-server/pkg/godoc"
	"golang.org/x/net/context"
)

func getSearchTool() (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("searchPackages",
			mcp.WithDescription("provide a query, search related golang packages from pkg.go.dev"),
			mcp.WithString("q",
				mcp.Required(),
				mcp.Description("query")),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			q, err := requiredParam[string](request, "q")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			search, err := godoc.Search(q)
			if err != nil {
				return nil, errors.WithMessage(err, "search failed")
			}
			marshal, err := json.Marshal(search.Packages)
			if err != nil {
				return nil, errors.Wrap(err, "marshal packages failed")
			}
			return mcp.NewToolResultText(string(marshal)), nil
		}
}
