package injector

import (
	"testing"
)

func TestModifyName(t *testing.T) {

	type testCase struct {
		name     string
		value    string
		hash     uint32
		expected string
	}

	testCases := []testCase{
		{name: "Empty", value: "", hash: 0, expected: ""},
		{name: "OnlyFile", value: "test", hash: 0x1234, expected: "test-1234"},
		{name: "FileWithExt", value: "command.com", hash: 0x5678, expected: "command-5678.com"},
		{name: "FileInDir", value: "/etc/passwd", hash: 0x12, expected: "/etc/passwd-12"},
		{name: "RegularFile", value: "/etc/sudoers.d", hash: 0x34, expected: "/etc/sudoers-34.d"},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if result := modifyName(tc.value, tc.hash); result != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, result)
			}
		})
	}
}

func TestCalculateHash(t *testing.T) {
	h, err := calculateHash("testdata/pasternak.txt")
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	if expected := uint32(0xfcc1be08); h != expected {
		t.Errorf("expected %0x, got %0x", expected, h)
	}
}
