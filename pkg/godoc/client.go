package godoc

import (
	"sync"

	"github.com/go-resty/resty/v2"
)

var client = sync.OnceValue(resty.New)

// 这个要考虑支持设置
func baseURL() string {
	return "https://pkg.go.dev"
}
