package tool

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pkg/errors"
	"github.com/yikakia/godoc-mcp-server/pkg/godoc"
)

type searchParams struct {
	Q string `json:"q" jsonschema:"query string"`
}

func GetSearchToolV2() mcp.ToolHandlerFor[searchParams, any] {
	return func(ctx context.Context, session *mcp.ServerSession, c *mcp.CallToolParamsFor[searchParams]) (*mcp.CallToolResultFor[any], error) {

		search, err := godoc.Search(c.Arguments.Q)
		if err != nil {
			return nil, errors.WithMessage(err, "search failed.")
		}
		marshal, err := json.Marshal(search.Packages)
		if err != nil {
			return nil, errors.Wrap(err, "marshal pkgs failed.")
		}
		return &mcp.CallToolResultFor[any]{
			Content: []mcp.Content{&mcp.TextContent{Text: string(marshal)}},
		}, nil
	}
}
