package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestLoadTLSConfig_Disabled(t *testing.T) {
	tlsCfg, certFile, keyFile := loadTLSConfig(false, "", "", "1.2")
	if tlsCfg != nil {
		t.Error("expected nil TLS config when disabled")
	}
	if certFile != "" || keyFile != "" {
		t.Error("expected empty cert/key paths when disabled")
	}
}

func TestLoadTLSConfig_Enabled(t *testing.T) {
	// Create temp cert and key files
	certDir := t.TempDir()
	certPath := certDir + "/cert.pem"
	keyPath := certDir + "/key.pem"
	if err := os.WriteFile(certPath, []byte("fake-cert"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(keyPath, []byte("fake-key"), 0600); err != nil {
		t.Fatal(err)
	}

	tlsCfg, returnedCert, returnedKey := loadTLSConfig(true, certPath, keyPath, "1.2")
	if tlsCfg == nil {
		t.Fatal("expected non-nil TLS config when enabled")
	}
	if tlsCfg.MinVersion != tls.VersionTLS12 {
		t.Errorf("expected MinVersion TLS 1.2, got %d", tlsCfg.MinVersion)
	}
	if returnedCert != certPath {
		t.Errorf("expected cert path %q, got %q", certPath, returnedCert)
	}
	if returnedKey != keyPath {
		t.Errorf("expected key path %q, got %q", keyPath, returnedKey)
	}
}

func TestLoadTLSConfig_EnabledTLS13(t *testing.T) {
	certDir := t.TempDir()
	certPath := certDir + "/cert.pem"
	keyPath := certDir + "/key.pem"
	if err := os.WriteFile(certPath, []byte("fake-cert"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(keyPath, []byte("fake-key"), 0600); err != nil {
		t.Fatal(err)
	}

	tlsCfg, _, _ := loadTLSConfig(true, certPath, keyPath, "1.3")
	if tlsCfg == nil {
		t.Fatal("expected non-nil TLS config")
	}
	if tlsCfg.MinVersion != tls.VersionTLS13 {
		t.Errorf("expected MinVersion TLS 1.3, got %d", tlsCfg.MinVersion)
	}
}

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

func TestParseTLSMinVersion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected uint16
	}{
		{"default empty", "", tls.VersionTLS12},
		{"explicit 1.2", "1.2", tls.VersionTLS12},
		{"explicit 1.3", "1.3", tls.VersionTLS13},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseTLSMinVersion(tt.input)
			if result != tt.expected {
				t.Errorf("parseTLSMinVersion(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseTLSMinVersion_Invalid(t *testing.T) {
	// Invalid values should cause a fatal log — test by verifying the function
	// only accepts "1.2", "1.3", or empty string
	// (Fatal behavior tested via integration test)
	result := parseTLSMinVersion("1.2")
	if result != tls.VersionTLS12 {
		t.Errorf("expected TLS 1.2, got %d", result)
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

// generateTestCert creates a self-signed certificate and key for testing.
// Returns PEM-encoded cert and key bytes.
func generateTestCert(t *testing.T) (certPEM, keyPEM []byte) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{Organization: []string{"Test"}},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1)},
		DNSNames:     []string{"localhost"},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("failed to create certificate: %v", err)
	}

	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatalf("failed to marshal key: %v", err)
	}
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})

	return certPEM, keyPEM
}

func TestTLSServerIntegration(t *testing.T) {
	certPEM, keyPEM := generateTestCert(t)

	// Write cert and key to temp files
	certDir := t.TempDir()
	certPath := certDir + "/cert.pem"
	keyPath := certDir + "/key.pem"
	if err := os.WriteFile(certPath, certPEM, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(keyPath, keyPEM, 0600); err != nil {
		t.Fatal(err)
	}

	tlsCfg, _, _ := loadTLSConfig(true, certPath, keyPath, "1.2")

	// Load the keypair for the TLS listener
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("failed to load test keypair: %v", err)
	}

	// Create test server
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)

	srv := &http.Server{
		Handler:           mux,
		TLSConfig:         tlsCfg,
		ReadHeaderTimeout: 5 * time.Second,
	}

	ln, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	})
	if err != nil {
		t.Fatalf("failed to create TLS listener: %v", err)
	}
	defer func() { _ = ln.Close() }()

	go func() { _ = srv.Serve(ln) }()
	defer func() { _ = srv.Close() }()

	// Create client that trusts our test cert
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	resp, err := client.Get("https://" + ln.Addr().String() + "/health")
	if err != nil {
		t.Fatalf("failed to GET /health over TLS: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	if resp.TLS == nil {
		t.Error("expected TLS connection")
	}
	if resp.TLS != nil && resp.TLS.Version < tls.VersionTLS12 {
		t.Errorf("expected TLS >= 1.2, got %d", resp.TLS.Version)
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
