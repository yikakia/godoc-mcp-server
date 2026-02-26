package godoc

import (
	"bytes"
	"compress/gzip"
	"fmt"

	"github.com/yikakia/cachalot/core/codec"
)

type GzipCodec struct{}

func (g *GzipCodec) Marshal(a any) ([]byte, error) {
	if a == nil {
		return nil, fmt.Errorf("GzipCodec requires an argument")
	}
	data, ok := a.([]byte)
	if !ok {
		return nil, fmt.Errorf("GzipCodec requires an argument as []byte")
	}

	return compress(data)
}

func (g *GzipCodec) Unmarshal(bytes []byte, a any) error {
	if a == nil {
		return fmt.Errorf("GzipCodec requires an argument")
	}
	data, ok := a.([]byte)
	if !ok {
		return fmt.Errorf("GzipCodec requires an argument as []byte")
	}
	data, err := decompress(data)
	if err != nil {
		return err
	}
	a = data
	return nil
}

var _ codec.Codec = (*GzipCodec)(nil)

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
