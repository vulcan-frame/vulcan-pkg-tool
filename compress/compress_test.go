package compress

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testWeakThreshold   = 1 << 10   // 1KB
	testStrongThreshold = 128 << 10 // 128KB
)

func TestMain(m *testing.M) {
	// 初始化测试配置
	Init(testWeakThreshold, testStrongThreshold)
	m.Run()
}

func TestCompressDecision(t *testing.T) {
	tests := []struct {
		name    string
		dataLen int
		want    bool
	}{
		{"BelowWeak", testWeakThreshold - 1, false},
		{"EqualWeak", testWeakThreshold, true},
		{"BetweenThresholds", testWeakThreshold + 1, true},
		{"EqualStrong", testStrongThreshold, true},
		{"AboveStrong", testStrongThreshold + 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make([]byte, tt.dataLen)
			_, compressed, _ := Compress(data)
			assert.Equal(t, compressed, tt.want)
		})
	}
}

func TestCompressDecompressCycle(t *testing.T) {
	testCases := []struct {
		name string
		data []byte
	}{
		{"Empty", []byte{}},
		{"Small", []byte("hello world")},
		{"Medium", bytes.Repeat([]byte{0x01}, testWeakThreshold)},
		{"Large", bytes.Repeat([]byte{0x01}, testStrongThreshold+1024)},
		{"Random", randBytes(2 * testStrongThreshold)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			compressed, didCompress, err := Compress(tc.data)
			require.Nil(t, err)
			if didCompress {
				decompressed, err := Decompress(compressed)
				require.Nil(t, err)
				assert.Equal(t, tc.data, decompressed)
			} else {
				assert.Equal(t, tc.data, compressed)
			}
		})
	}
}

func TestErrorConditions(t *testing.T) {
	t.Run("InvalidDecompressData", func(t *testing.T) {
		_, err := Decompress([]byte{0x00, 0x01, 0x02})
		assert.NotNil(t, err)
	})

	t.Run("NilInput", func(t *testing.T) {
		t.Run("Compress", func(t *testing.T) {
			compressed, didCompress, err := Compress(nil)
			assert.Nil(t, err)
			assert.Equal(t, []byte{}, compressed)
			assert.Equal(t, false, didCompress)
		})

		t.Run("Decompress", func(t *testing.T) {
			decompressed, err := Decompress(nil)
			assert.Nil(t, err)
			assert.Equal(t, []byte{}, decompressed)
		})
	})
}

func TestConcurrentSafety(t *testing.T) {
	var wg sync.WaitGroup
	const goroutines = 10

	// Test concurrent Init and Compress
	wg.Add(goroutines * 2)
	for i := 0; i < goroutines; i++ {
		go func(v int) {
			defer wg.Done()
			Init(v*testWeakThreshold, v*testStrongThreshold)
		}(i)

		go func() {
			defer wg.Done()
			data := randBytes(testStrongThreshold * 2)
			_, _, _ = Compress(data)
		}()
	}
	wg.Wait()
}

func BenchmarkCompress(b *testing.B) {
	sizes := []struct {
		name string
		size int
	}{
		{"64B", 64},
		{"1KB", 1 << 10},
		{"512KB", testWeakThreshold},
		{"1MB", testStrongThreshold},
		{"4MB", 4 << 20},
	}

	for _, size := range sizes {
		data := randBytes(size.size)
		b.Run(size.name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(size.size))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _, _ = Compress(data)
			}
		})
	}
}

func BenchmarkDecompress(b *testing.B) {
	sizes := []struct {
		name string
		size int
	}{
		{"64B", 64},
		{"1KB", 1 << 10},
		{"512KB", testWeakThreshold},
		{"1MB", testStrongThreshold},
		{"4MB", 4 << 20},
	}

	for _, size := range sizes {
		data := randBytes(size.size)
		compressed, _, _ := Compress(data)
		b.Run(size.name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(compressed)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = Decompress(compressed)
			}
		})
	}
}

func randBytes(n int) []byte {
	data := make([]byte, n)
	_, _ = rand.Read(data) // 对于测试来说，错误可以忽略
	return data
}

func TestCompressionEfficiency(t *testing.T) {
	type testData struct {
		ID      int64
		Name    string
		Tags    []string
		Value   float64
		Enabled bool
	}

	generateData := func(sizeKB int) []byte {
		data := make([]byte, 0, sizeKB<<10)
		buf := bytes.NewBuffer(data)

		// 生成结构化数据
		for i := 0; buf.Len() < sizeKB<<10; i++ {
			item := testData{
				ID:      int64(i),
				Name:    fmt.Sprintf("item-%d", i),
				Tags:    []string{"tag1", "tag2", "tag3"},
				Value:   rand.Float64(),
				Enabled: i%2 == 0,
			}
			// 使用JSON模拟proto序列化
			b, _ := json.Marshal(item)
			buf.Write(b)
		}
		return buf.Bytes()[:sizeKB<<10] // 精确控制大小
	}

	testCases := []struct {
		name         string
		sizeKB       int
		wantMinRatio float64 // 预期的最低压缩率
	}{
		{"SmallData(10KB)", 12, 0.3},
		{"MediumData(100KB)", 100, 0.2},
		{"LargeData(1MB)", 1024, 0.2},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data := generateData(tc.sizeKB)
			origSize := len(data)

			// 压缩测试
			startCompress := time.Now()
			compressed, didCompress, err := Compress(data)
			compressTime := time.Since(startCompress)
			require.NoError(t, err)
			require.True(t, didCompress)

			// 解压测试
			startDecompress := time.Now()
			decompressed, err := Decompress(compressed)
			decompressTime := time.Since(startDecompress)
			require.NoError(t, err)

			// 验证数据完整性
			assert.Equal(t, data, decompressed)

			// 计算指标
			compressedSize := len(compressed)
			ratio := float64(compressedSize) / float64(origSize)

			// 输出性能报告
			t.Logf("Original size: %d KB", origSize>>10)
			t.Logf("Compressed size: %d KB (%.2f%%)",
				compressedSize>>10, ratio*100)
			t.Logf("Compress time: %s (%.2f MB/s)",
				compressTime,
				float64(origSize)/(compressTime.Seconds()*(1<<20)))
			t.Logf("Decompress time: %s (%.2f MB/s)",
				decompressTime,
				float64(origSize)/(decompressTime.Seconds()*(1<<20)))

			// 验证压缩率
			assert.Less(t, ratio, tc.wantMinRatio,
				"compression ratio too high, expected < %.2f, got %.2f",
				tc.wantMinRatio, ratio)
		})
	}
}
