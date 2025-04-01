package dnsx

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/IxBahy/ASM/internal/scanners"
	"github.com/IxBahy/ASM/pkg/utils/extractor"
	"github.com/projectdiscovery/dnsx/libs/dnsx"
)

type DNSxScanner struct {
	*scanners.BaseScanner
}

type DNSResult struct {
	Domain    string   `json:"domain"`
	IPs       []string `json:"ips"`
	TimeStamp string   `json:"timestamp"`
}

func NewDNSxScanner() *DNSxScanner {
	config := scanners.ScannerConfig{
		Name:           "dnsx",
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

	s := &DNSxScanner{
		BaseScanner: base,
	}
	return s
}

func (s *DNSxScanner) Setup() error {
	return nil
}

func (s *DNSxScanner) IsInstalled() bool {
	return true
}

func (s *DNSxScanner) RegisterInstallationStats() error {
	return nil
}

func (s *DNSxScanner) Scan(target string) (scanners.ScannerResult, error) {
	result := scanners.ScannerResult{
		Data:   []string{},
		Errors: []string{},
	}

	domain := extractor.ExtractDomain(target)

	options := dnsx.DefaultOptions

	dnsClient, err := dnsx.New(options)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("failed to create dnsx client: %v", err))
		return result, err
	}

	ips, err := dnsClient.Lookup(domain)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("lookup error: %v", err))

		ips = []string{}
	}

	dnsResult := DNSResult{
		Domain:    domain,
		IPs:       ips,
		TimeStamp: time.Now().Format(time.RFC3339),
	}

	jsonData, err := json.MarshalIndent(dnsResult, "", "  ")
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("json encoding error: %v", err))
		return result, err
	}

	result.Data = append(result.Data, string(jsonData))
	return result, nil
}
