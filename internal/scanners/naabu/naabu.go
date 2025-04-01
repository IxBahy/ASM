package naabu

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IxBahy/ASM/internal/scanners"
	"github.com/IxBahy/ASM/pkg/utils/extractor"
	"github.com/projectdiscovery/goflags"
	naabuResult "github.com/projectdiscovery/naabu/v2/pkg/result"
	"github.com/projectdiscovery/naabu/v2/pkg/runner"
)

type NaabuScanner struct {
	*scanners.BaseScanner
}

type PortScanResult struct {
	Host      string   `json:"host"`
	IP        string   `json:"ip"`
	Ports     []string `json:"ports"`
	TimeStamp string   `json:"timestamp"`
}

func NewNaabuScanner() *NaabuScanner {
	config := scanners.ScannerConfig{
		Name:           "naabu",
		Version:        "embedded",
		ExecutablePath: "",
		Base_Command:   "",
	}

	base := &scanners.BaseScanner{
		Config: config,
		InstallState: scanners.InstallationState{
			Installed: true,
			Version:   "2.1.0",
		},
	}

	s := &NaabuScanner{
		BaseScanner: base,
	}
	return s
}

func (s *NaabuScanner) Setup() error {
	return nil
}

func (s *NaabuScanner) IsInstalled() bool {
	return true
}

func (s *NaabuScanner) RegisterInstallationStats() error {
	return nil
}

func (s *NaabuScanner) Scan(target string) (scanners.ScannerResult, error) {
	result := scanners.ScannerResult{
		Data:   []string{},
		Errors: []string{},
	}

	host := extractor.ExtractDomain(target)

	var ports []string

	options := runner.Options{
		Host:    goflags.StringSlice{host},
		Ports:   "80,443,8080,8443",
		Silent:  true,
		NoColor: true,
		JSON:    true,

		OnResult: func(hr *naabuResult.HostResult) {
			if len(hr.Ports) > 0 {
				for _, port := range hr.Ports {
					ports = append(ports, fmt.Sprintf("%d", port.Port))
				}
			}
		},
	}

	naabuRunner, err := runner.NewRunner(&options)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("failed to initialize naabu: %v", err))
		return result, err
	}
	defer naabuRunner.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	err = naabuRunner.RunEnumeration(ctx)

	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("port scanning error: %v", err))
		return result, err
	}

	scanResult := PortScanResult{
		Host:      host,
		IP:        "",
		Ports:     ports,
		TimeStamp: time.Now().Format(time.RFC3339),
	}

	jsonData, err := json.MarshalIndent(scanResult, "", "  ")
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("json encoding error: %v", err))
		return result, err
	}

	result.Data = append(result.Data, string(jsonData))
	return result, nil
}
