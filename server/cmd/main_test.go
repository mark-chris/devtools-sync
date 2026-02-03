package main

import (
	"testing"
)

func TestParseMaxBodySize_Default(t *testing.T) {
	result := parseMaxBodySize("")
	expected := int64(10 * 1024 * 1024) // 10MB

	if result != expected {
		t.Errorf("parseMaxBodySize(\"\") = %d, want %d", result, expected)
	}
}

func TestParseMaxBodySize_Bytes(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"1024", 1024},
		{"1048576", 1048576},
		{"0", 0},
		{"100", 100},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseMaxBodySize(tt.input)
			if result != tt.expected {
				t.Errorf("parseMaxBodySize(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseMaxBodySize_KB(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"1KB", 1024},
		{"10KB", 10 * 1024},
		{"512KB", 512 * 1024},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseMaxBodySize(tt.input)
			if result != tt.expected {
				t.Errorf("parseMaxBodySize(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseMaxBodySize_MB(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"1MB", 1024 * 1024},
		{"10MB", 10 * 1024 * 1024},
		{"50MB", 50 * 1024 * 1024},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseMaxBodySize(tt.input)
			if result != tt.expected {
				t.Errorf("parseMaxBodySize(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseMaxBodySize_GB(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"1GB", 1024 * 1024 * 1024},
		{"2GB", 2 * 1024 * 1024 * 1024},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseMaxBodySize(tt.input)
			if result != tt.expected {
				t.Errorf("parseMaxBodySize(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseMaxBodySize_Invalid(t *testing.T) {
	tests := []string{
		"invalid",
		"abc",
		"10XB",
		"",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			result := parseMaxBodySize(input)
			// Invalid input should return default (10MB)
			if input != "" && result != 10*1024*1024 {
				t.Errorf("parseMaxBodySize(%q) = %d, want default 10MB", input, result)
			}
		})
	}
}

func TestParseMaxBodySize_EdgeCases(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"0MB", 0},
		{"0KB", 0},
		{"1", 1},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseMaxBodySize(tt.input)
			if result != tt.expected {
				t.Errorf("parseMaxBodySize(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}
