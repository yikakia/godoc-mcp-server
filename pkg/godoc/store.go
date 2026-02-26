package godoc

import (
	"sync"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/yikakia/cachalot/core/cache"
	store_ristretto "github.com/yikakia/cachalot/stores/ristretto"
)

var store = sync.OnceValue(func() cache.Store {
	return initStore()
})

func initStore() cache.Store {
	client, err := ristretto.NewCache(&ristretto.Config[string, any]{
		NumCounters: 1 << 10,
		MaxCost:     1 << 20,
		BufferItems: 64,
	})
	if err != nil {
		panic(err)
	}
	store := store_ristretto.New(client, store_ristretto.WithStoreName("basic-ristretto"))
	return store
}
