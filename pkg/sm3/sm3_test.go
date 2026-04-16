package sm3

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSM3StandardVectors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			// 示例 1: "abc" (出自官方标准 GM/T 0004-2012)
			input:    "abc",
			expected: "66c7f0f462eeedd9d1f2d46bdc10e4e24167c4875cf2f7a2297da02b8f4ba8e0",
		},
		{
			// 示例 2: 长度为 64 字节（512比特）的重复字符 "abcd..."
			input:    "abcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcd",
			expected: "debe9ff92275b8a138604889c18e5a4d6fdb70e5387e5765293dcba39c0c5732",
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Input: %s", tt.input), func(t *testing.T) {
			hash := Sum([]byte(tt.input))
			actual := hex.EncodeToString(hash[:])
			assert.Equal(t, tt.expected, actual, "Hash value should match standard SM3 result")
		})
	}
}

func TestSM3IncrementalWrite(t *testing.T) {
	// 测试分次写入 Write 是否与一次性写入 Sum 结果一致
	msg := "Welcome to Higress WASM Plugin Development with SM3 Support"
	
	// 一次性写入
	expectedHash := Sum([]byte(msg))
	
	// 分次写入
	d := New()
	d.Write([]byte(msg[:10]))
	d.Write([]byte(msg[10:20]))
	d.Write([]byte(msg[20:]))
	actualHash := d.Sum(nil)
	
	assert.Equal(t, hex.EncodeToString(expectedHash[:]), hex.EncodeToString(actualHash), "Incremental write should produce same result")
}

func BenchmarkSM3(b *testing.B) {
	data := []byte("A quick brown fox jumps over the lazy dog. SM3 is fast and secure.")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Sum(data)
	}
}
