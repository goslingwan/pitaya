package compression

import (
	"bytes"
	"compress/zlib"
	"io/ioutil"
	"sync"
)

var (
	// 公共对象池,更极致的优化可以建多个池
	zlibPool = sync.Pool{New: func() interface{} {
		buf := new(bytes.Buffer)
		return zlib.NewWriter(buf)
	}}
	bufPool = sync.Pool{New: func() interface{} {
		return new(bytes.Buffer)
	}}
)

func init() {
	for i := 0; i < 5000; i++ {
		buf := new(bytes.Buffer)
		bufPool.Put(buf)
		z := zlib.NewWriter(buf)
		zlibPool.Put(z)
	}
}

func DeflateData(data []byte) ([]byte, error) {
	buf := bufPool.Get().(*bytes.Buffer)
	z := zlibPool.Get().(*zlib.Writer)
	z.Reset(buf)
	defer func() {
		// 归还buff
		buf.Reset()
		bufPool.Put(buf)
		// 归还Writer
		zlibPool.Put(z)
	}()
	z.Reset(buf)
	_, err := z.Write(data)
	if err != nil {
		return nil, err
	}
	z.Close()
	return buf.Bytes(), nil
}

func InflateData(data []byte) ([]byte, error) {
	zr, err := zlib.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	defer zr.Close()

	return ioutil.ReadAll(zr)
}

func IsCompressed(data []byte) bool {
	return len(data) > 2 &&
		(
		// zlib
		(data[0] == 0x78 &&
			(data[1] == 0x9C ||
				data[1] == 0x01 ||
				data[1] == 0xDA ||
				data[1] == 0x5E)) ||
			// gzip
			(data[0] == 0x1F && data[1] == 0x8B))
}
