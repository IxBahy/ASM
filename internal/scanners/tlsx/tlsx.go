package tlsx

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/IxBahy/ASM/internal/scanners"
	"github.com/IxBahy/ASM/pkg/utils/extractor"
)

type TLSXScanner struct {
	*scanners.BaseScanner
}

// TLSResult represents the result of a TLS scan
type TLSResult struct {
	Host                    string   `json:"host"`
	Port                    string   `json:"port"`
	TLSVersion              string   `json:"tls_version"`
	CipherSuite             string   `json:"cipher_suite"`
	CertificateExpiry       string   `json:"certificate_expiry"`
	CertificateIssuer       string   `json:"certificate_issuer"`
	CertificateSubject      string   `json:"certificate_subject"`
	CertificateSerialNumber string   `json:"certificate_serial_number"`
	DNSNames                []string `json:"dns_names"`
	IsValid                 bool     `json:"is_valid"`
	ValidationErrors        []string `json:"validation_errors,omitempty"`
}

func NewTLSXScanner() *TLSXScanner {
	config := scanners.ScannerConfig{
		Name:           "tlsx",
		Version:        "embedded",
		ExecutablePath: "",
		Base_Command:   "",
	}

	base := &scanners.BaseScanner{
		Config: config,
		InstallState: scanners.InstallationState{
			Installed: true,
			Version:   "1.0.0",
		},
	}

	s := &TLSXScanner{
		BaseScanner: base,
	}
	return s
}

func (s *TLSXScanner) Setup() error {
	return nil
}

func (s *TLSXScanner) IsInstalled() bool {
	return true
}

func (s *TLSXScanner) RegisterInstallationStats() error {
	return nil
}

func (s *TLSXScanner) Scan(target string) (scanners.ScannerResult, error) {
	result := scanners.ScannerResult{
		Data:   []string{},
		Errors: []string{},
	}

	host, port := extractor.ExtractHostAndPort(target)
	if port == "" {
		port = "443"
	}

	conn, err := tls.DialWithDialer(
		&net.Dialer{Timeout: 10 * time.Second},
		"tcp",
		fmt.Sprintf("%s:%s", host, port),
		&tls.Config{
			InsecureSkipVerify: true,
		},
	)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("connection error: %v", err))
		return result, err
	}
	defer conn.Close()

	state := conn.ConnectionState()

	tlsResult := TLSResult{
		Host:        host,
		Port:        port,
		TLSVersion:  getTLSVersionString(state.Version),
		CipherSuite: tls.CipherSuiteName(state.CipherSuite),
	}

	if len(state.PeerCertificates) > 0 {
		cert := state.PeerCertificates[0]
		tlsResult.CertificateExpiry = cert.NotAfter.Format(time.RFC3339)
		tlsResult.CertificateSerialNumber = cert.SerialNumber.String()

		// Format the issuer and subject using direct field access
		tlsResult.CertificateIssuer = formatCertName(cert.Issuer.String())
		tlsResult.CertificateSubject = formatCertName(cert.Subject.String())
		tlsResult.DNSNames = cert.DNSNames

		// Validate certificate
		opts := x509.VerifyOptions{
			DNSName: host,
		}
		_, err := cert.Verify(opts)
		tlsResult.IsValid = (err == nil)
		if err != nil {
			errorStr := err.Error()
			tlsResult.ValidationErrors = []string{errorStr}

			// Split into multiple errors if there are multiple issues
			if strings.Contains(errorStr, "; ") {
				tlsResult.ValidationErrors = strings.Split(errorStr, "; ")
			}
		}
	}

	jsonData, err := json.MarshalIndent(tlsResult, "", "  ")
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("json encoding error: %v", err))
		return result, err
	}

	result.Data = append(result.Data, string(jsonData))
	return result, nil
}

func getTLSVersionString(version uint16) string {
	switch version {
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	default:
		return fmt.Sprintf("Unknown (0x%04X)", version)
	}
}

func formatCertName(name string) string {
	return name
}
