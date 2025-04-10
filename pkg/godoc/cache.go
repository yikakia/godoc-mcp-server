package godoc

import (
	"bytes"
	"compress/gzip"
	"log"
	"sync"

	"github.com/coocood/freecache"
	"github.com/pkg/errors"
)

// cacheSize 50M
var cacheSize = 50 * 1024 * 1024

// cacheTTL600s 10 minutes
var cacheTTL600s = 600

var cache = sync.OnceValue(func() *freecache.Cache {
	c := freecache.NewCache(cacheSize)

	return c
})

func getWithFn(key string, fn func() ([]byte, error)) ([]byte, error) {
	v, exist, err := get(key)
	if err != nil {
		return v, err
	}
	if exist {
		return v, nil
	}
	// not exist
	v, err = fn()
	if err != nil {
		return v, err
	}

	go func() {
		defer func() {
			if t := recover(); t != nil {
				log.Println("recovered from ", t)
			}
		}()
		data, err := compress(v)
		if err != nil {
			log.Println(err)
			return
		}

		// set to cache
		err = cache().Set([]byte(key), data, cacheTTL600s)
		if errors.Is(err, freecache.ErrLargeEntry) {
			log.Print("large entry for key", key)
			return
		}
		if err != nil {
			log.Println(err)
		}
	}()

	return v, nil
}

func get(key string) ([]byte, bool, error) {
	data, err := cache().Get([]byte(key))
	if errors.Is(err, freecache.ErrNotFound) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, errors.WithStack(err)
	}
	data, err = decompress(data)
	if err != nil {
		return nil, false, errors.WithStack(err)
	}

	return data, true, nil
}

func compress(data []byte) ([]byte, error) {
	var b bytes.Buffer
	gz, err := gzip.NewWriterLevel(&b, gzip.BestCompression)
	if err != nil {
		return nil, err
	}
	_, err = gz.Write(data)
	if err != nil {
		return nil, err
	}
	err = gz.Close()
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func decompress(data []byte) ([]byte, error) {
	b := bytes.NewBuffer(data)
	gz, err := gzip.NewReader(b)
	if err != nil {
		return nil, err
	}
	defer gz.Close()
	var res bytes.Buffer
	_, err = res.ReadFrom(gz)
	if err != nil {
		return nil, err
	}
	return res.Bytes(), nil
}
