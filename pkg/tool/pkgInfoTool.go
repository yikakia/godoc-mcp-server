package tool

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pkg/errors"
	"github.com/yikakia/godoc-mcp-server/pkg/godoc"
)

type GetPkgInfoParams struct {
	// TODO add the description to the filed pkgName and needURL

	// pkgName
	// package name for search. if use searchPackages before, and user want to get the subpackage info. you should plus
	// them for example, when user query mcp, and it return packageName: github.com/mark3labs/mcp-go/mcp and subpackage
	// client, then if user want to get the client package info, you should set the packageName to
	// github.com/mark3labs/mcp-go/mcp/client rather than client
	PkgName string `json:"pkgName" jsonschema:"the package name user search"`
	// default is false. if it`s true, will return the url of the definition of the package`s consts,types,functions,
	// variables,subpackages. only when user need it, set it
	NeedURL bool `json:"needURL" jsonschema:"if user need the link to the definition"`
}

func GetPkgInfoTool() mcp.ToolHandlerFor[GetPkgInfoParams, *godoc.PackageDocument] {
	return func(ctx context.Context, c *mcp.CallToolRequest, input GetPkgInfoParams) (*mcp.CallToolResult, *godoc.PackageDocument, error) {
		pkgDoc, err := godoc.GetPackageDocument(godoc.GetPackageRequest{
			PackageName: input.PkgName,
			NeedURL:     input.NeedURL,
		})
		if err != nil {
			return nil, nil, errors.WithMessage(err, "get pkg info failed")
		}

		return nil, pkgDoc, nil
	}
}
