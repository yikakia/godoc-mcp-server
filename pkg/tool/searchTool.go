package tool

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pkg/errors"
	"github.com/yikakia/godoc-mcp-server/pkg/godoc"
)

type searchParams struct {
	Q string `json:"q" jsonschema:"query string"`
}

func GetSearchTool() mcp.ToolHandlerFor[searchParams, *godoc.SearchResult] {
	return func(ctx context.Context, c *mcp.CallToolRequest, input searchParams) (*mcp.CallToolResult, *godoc.SearchResult, error) {

		search, err := godoc.Search(input.Q)
		if err != nil {
			return nil, nil, errors.WithMessage(err, "search failed.")
		}

		return nil, search, nil
	}
}
