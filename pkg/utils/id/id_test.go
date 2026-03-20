package id

import (
	"testing"
	"time"
)

func TestSnowflake(t *testing.T) {
	// 测试初始化
	err := Init(1)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// 重复初始化应该返回错误（sync.Once 保证只执行一次）
	err = Init(2)
	// 注意：由于使用 sync.Once，第二次不会返回错误，只是不执行

	// 测试生成 ID
	id1, err := Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	id2, err := Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// ID 应该是递增的
	if id2 <= id1 {
		t.Errorf("IDs should be increasing: %d >= %d", id2, id1)
	}

	// 提取时间戳
	ts := ExtractTimestamp(id1)
	if ts == 0 {
		t.Error("Extracted timestamp should not be 0")
	}

	// 提取机器 ID
	mid := ExtractMachineID(id1)
	if mid != 1 {
		t.Errorf("Machine ID should be 1, got %d", mid)
	}

	// 提取生成时间
	genTime := TimeFromID(id1)
	if genTime.After(time.Now()) {
		t.Error("Generated time should not be in the future")
	}
}

func TestEncodeDecode(t *testing.T) {
	testCases := []int64{
		0,
		1,
		100,
		1000,
		123456789,
		9223372036854775807, // Max int64
		-1,
		-100,
		-9223372036854775808, // Min int64
	}

	for _, original := range testCases {
		encoded := Itoa(original)
		decoded := Atoi(encoded)

		if decoded != original {
			t.Errorf("Encode/Decode failed for %d: got %d", original, decoded)
		}
	}
}

func TestIsValidIDString(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
	}{
		{"", false},
		{"0", true},
		{"abc", true},
		{"ABC", true},
		{"123", true},
		{"-abc", true},
		{"-", false},
		{"abc!", false},
		{"abc def", false},
	}

	for _, tc := range testCases {
		result := IsValidIDString(tc.input)
		if result != tc.expected {
			t.Errorf("IsValidIDString(%q) = %v, expected %v", tc.input, result, tc.expected)
		}
	}
}

func BenchmarkGenerate(b *testing.B) {
	Init(1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Generate()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkItoa(b *testing.B) {
	id := int64(123456789012345)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Itoa(id)
	}
}

func BenchmarkAtoi(b *testing.B) {
	s := Itoa(123456789012345)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Atoi(s)
	}
}
