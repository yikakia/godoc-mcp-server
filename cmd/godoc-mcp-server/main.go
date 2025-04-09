package main

import (
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/server"
)

var (
	version = "1.0.0"
)

func main() {
	mcpServer := server.NewMCPServer(
		"godoc-mcp-server",
		version,
	)
	mcpServer.AddTool(getSearchTool())
	mcpServer.AddTool(getPackageInfoTool())

	_, _ = fmt.Fprintf(os.Stderr, "godoc-mcp-server running on stdio...\n")
	err := server.ServeStdio(mcpServer)
	if err != nil {
		panic(err)
	}
}
