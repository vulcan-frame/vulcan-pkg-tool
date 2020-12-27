package compress

import (
	"bytes"
	"compress/zlib"
	"sync"

	"github.com/pkg/errors"
)

var (
	compressMutex         sync.RWMutex
	defaultWeakCompress   = 10 << 10  // 10KB
	defaultStrongCompress = 512 << 10 // 512KB
	defaultWeakLevel      = zlib.BestSpeed
	defaultStrongLevel    = zlib.DefaultCompression
)

var (
	compressBufferPool = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
	decompressBufferPool = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
)

// Init init compress params
// weak: weak compress threshold, compress when data length is greater than this value
// strong: strong compress threshold, use higher compression rate when data length is greater than this value
func Init(weak, strong int) {
	compressMutex.Lock()
	defer compressMutex.Unlock()

	if weak > 0 {
		defaultWeakCompress = weak
	}
	if strong > 0 {
		defaultStrongCompress = strong
	}
}

// Compress auto select compress strategy based on data length
// return compressed data, whether compression is performed, error info
func Compress(data []byte) ([]byte, bool, error) {
	dataLen := len(data)
	if dataLen == 0 {
		return []byte{}, false, nil
	}

	compressMutex.RLock()
	weakThreshold := defaultWeakCompress
	strongThreshold := defaultStrongCompress
	compressMutex.RUnlock()

	if dataLen < weakThreshold {
		return data, false, nil
	}

	level := defaultWeakLevel
	if dataLen >= strongThreshold {
		level = defaultStrongLevel
	}

	compressed, err := zlibCompress(data, level)
	if err != nil {
		return nil, false, errors.Wrap(err, "compression failed")
	}
	return compressed, true, nil
}

// Decompress decompress data
func Decompress(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return []byte{}, nil
	}

	decompressed, err := zlibDecompress(data)
	if err != nil {
		return nil, errors.Wrap(err, "decompression failed")
	}
	return decompressed, nil
}

func zlibCompress(data []byte, level int) ([]byte, error) {
	if level < zlib.BestSpeed || level > zlib.BestCompression {
		level = zlib.DefaultCompression
	}

	buffer := compressBufferPool.Get().(*bytes.Buffer)
	defer func() {
		buffer.Reset()
		compressBufferPool.Put(buffer)
	}()

	writer, err := zlib.NewWriterLevel(buffer, level)
	if err != nil {
		return nil, errors.Wrapf(err, "create zlib writer failed (level %d)", level)
	}

	if _, err := writer.Write(data); err != nil {
		writer.Close()
		return nil, errors.Wrap(err, "write to compressor failed")
	}

	if err := writer.Close(); err != nil {
		return nil, errors.Wrap(err, "close compressor failed")
	}

	return buffer.Bytes(), nil
}

func zlibDecompress(data []byte) ([]byte, error) {
	reader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, errors.Wrap(err, "create zlib reader failed")
	}
	defer reader.Close()

	buffer := decompressBufferPool.Get().(*bytes.Buffer)
	defer func() {
		buffer.Reset()
		decompressBufferPool.Put(buffer)
	}()

	if _, err := buffer.ReadFrom(reader); err != nil {
		return nil, errors.Wrap(err, "read from decompressor failed")
	}
	return buffer.Bytes(), nil
}
