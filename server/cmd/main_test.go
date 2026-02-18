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

func TestParseCORSOrigins_Empty(t *testing.T) {
	result := parseCORSOrigins("")
	if len(result) != 0 {
		t.Errorf("Expected empty slice, got %v", result)
	}
}

func TestParseCORSOrigins_Single(t *testing.T) {
	result := parseCORSOrigins("http://localhost:5173")
	if len(result) != 1 || result[0] != "http://localhost:5173" {
		t.Errorf("Expected [http://localhost:5173], got %v", result)
	}
}

func TestParseCORSOrigins_Multiple(t *testing.T) {
	result := parseCORSOrigins("http://localhost:5173,https://dashboard.example.com")
	if len(result) != 2 {
		t.Fatalf("Expected 2 origins, got %d", len(result))
	}
	if result[0] != "http://localhost:5173" {
		t.Errorf("Expected first origin 'http://localhost:5173', got %q", result[0])
	}
	if result[1] != "https://dashboard.example.com" {
		t.Errorf("Expected second origin 'https://dashboard.example.com', got %q", result[1])
	}
}

func TestParseCORSOrigins_WhitespaceHandling(t *testing.T) {
	result := parseCORSOrigins(" http://localhost:5173 , https://dashboard.example.com ")
	if len(result) != 2 {
		t.Fatalf("Expected 2 origins, got %d", len(result))
	}
	if result[0] != "http://localhost:5173" {
		t.Errorf("Expected trimmed first origin, got %q", result[0])
	}
	if result[1] != "https://dashboard.example.com" {
		t.Errorf("Expected trimmed second origin, got %q", result[1])
	}
}
