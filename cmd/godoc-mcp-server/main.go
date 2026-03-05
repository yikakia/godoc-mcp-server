package main

import (
	"context"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	s := initServer()

	err := s.Run(context.Background(), &mcp.StdioTransport{})
	if err != nil {
		log.Fatal("unknown err, will exit. err:", err)
	}
}
