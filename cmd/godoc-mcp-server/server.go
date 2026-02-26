package main

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yikakia/godoc-mcp-server/pkg/tool"
)

var (
	name    = "godoc-mcp-server"
	version = "v1.0.1"
)

func initServer() *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    name,
		Version: version,
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Description: "provide a golang package name,get package consts,types,functions,variables," +
			"subpackages and how to use it. If return is null then means cannot find the package by the given name",
		Name: "getPackageInfo",
	}, tool.GetPkgInfoTool())

	mcp.AddTool(server, &mcp.Tool{
		Description: "provide a query, search related golang packages from pkg.go.dev include " +
			"name, path, synopsis, go doc url, imported by how many packages, subpackages in this package " +
			"the path is the package full name. if want to use getPackageInfo. llm should pass the path as " +
			"packageName to getPackageInfo. If return is null then means cannot find the package by the given name",
		Name: "searchPackages",
	}, tool.GetSearchToolV2())

	return server
}
