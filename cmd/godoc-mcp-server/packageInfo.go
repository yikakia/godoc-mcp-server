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
			mcp.WithDescription("provide a golang package name,get package consts,types,functions,variables,"+
				"subpackages and how to use it"),
			mcp.WithString("packageName",
				mcp.Required(),
				mcp.Description("package name for search. if use searchPackages before, and user want to get the "+
					"subpackage info. you should plus them for example, when user query mcp, and it return packageName: "+
					"github.com/mark3labs/mcp-go/mcp and subpackage client, then if user want to get the client package info, "+
					"you should set the packageName to github.com/mark3labs/mcp-go/mcp/client rather than client"),
			),
			mcp.WithBoolean("needURL",
				mcp.Description("default is false. if it`s true, will return the url of the definition of the "+
					"package`s consts,types,functions,variables,subpackages. only when user need it, set it"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			pkgName, err := requiredParam[string](request, "packageName")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			needURL, _, err := OptionalParamOK[bool](request, "needURL")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			pkgDoc, err := godoc.GetPackageDocument(godoc.GetPackageRequest{
				PackageName: pkgName,
				NeedURL:     needURL,
			})
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
