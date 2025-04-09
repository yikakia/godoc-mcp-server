package main

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/pkg/errors"
	"github.com/yikakia/godoc-mcp-server/pkg/godoc"
)

func getPackageInfoTool() (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("getPackageInfo",
			mcp.WithDescription("provide a golang package name,get package consts,types,functions,variables from pkg.go.dev"),
			mcp.WithString("packageName",
				mcp.Required(),
				mcp.Description("package name for search")),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			pkgName, err := requiredParam[string](request, "packageName")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			pkgDoc, err := godoc.GetPackageDocument(pkgName)
			if err != nil {
				return nil, errors.WithMessage(err, "get pkg info failed")
			}
			marshal, err := json.Marshal(pkgDoc)
			if err != nil {
				return nil, errors.Wrap(err, "marshal pkgDoc failed")
			}
			return mcp.NewToolResultText(string(marshal)), nil
		}
}
