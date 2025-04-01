package aiodnsbrute

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/IxBahy/ASM/internal/scanners"
	"github.com/IxBahy/ASM/pkg/client"
	"github.com/IxBahy/ASM/pkg/utils/extractor"
)

type AioDNSBruteScanner struct {
	*scanners.BaseScanner
	installClient client.ToolInstaller
}

type DNSBruteResult struct {
	Domain     string   `json:"domain"`
	Subdomains []string `json:"subdomains"`
	TimeStamp  string   `json:"timestamp"`
}

func NewAioDNSBruteScanner() *AioDNSBruteScanner {
	config := scanners.ScannerConfig{
		Name:             "aiodnsbrute",
		Version:          "latest",
		ExecutablePath:   "/usr/local/bin/aiodnsbrute",
		Base_Command:     "aiodnsbrute",
		InstallationType: client.InstallationTypePython,
	}

	base := &scanners.BaseScanner{
		Config: config,
		InstallState: scanners.InstallationState{
			Installed: false,
			Version:   "",
		},
	}

	s := &AioDNSBruteScanner{
		BaseScanner: base,
	}
	s.InstallState.Installed = s.IsInstalled()
	return s
}

func (s *AioDNSBruteScanner) Setup() error {
	if s.IsInstalled() {
		return nil
	}

	//
	installArgs := []string{"aiodnsbrute", s.Config.Version}

	var err error
	s.installClient, err = client.ClientFactory(s.Config.InstallationType, installArgs, 5)
	if err != nil {
		return fmt.Errorf("failed to create install client: %w", err)
	}

	if err := s.installClient.InstallTool(); err != nil {
		return fmt.Errorf("failed to install aiodnsbrute: %w", err)
	}

	if path, err := exec.LookPath("aiodnsbrute"); err == nil {
		s.Config.ExecutablePath = path
	}

	return s.RegisterInstallationStats()
}

func (s *AioDNSBruteScanner) Scan(target string) (scanners.ScannerResult, error) {

	result := scanners.ScannerResult{
		Data:   []string{},
		Errors: []string{},
	}

	domain := extractor.ExtractDomain(target)

	cmdParts := strings.Fields(s.Config.Base_Command)
	cmdParts = append(cmdParts, domain, "--output", "json", "-f", "aiodnsbrute.json")

	cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
	output, err := cmd.CombinedOutput()

	if err != nil && len(output) == 0 {
		result.Errors = append(result.Errors, fmt.Sprintf("command error: %v", err))
		return result, err
	}

	subdomains := []string{}

	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var entry struct {
			Name  string `json:"name"`
			Type  string `json:"type"`
			Value string `json:"value"`
		}

		if err := json.Unmarshal([]byte(line), &entry); err == nil && entry.Name != "" {
			if !contains(subdomains, entry.Name) {
				subdomains = append(subdomains, entry.Name)
			}
		}
	}

	dnsResult := DNSBruteResult{
		Domain:     domain,
		Subdomains: subdomains,
		TimeStamp:  time.Now().Format(time.RFC3339),
	}

	jsonData, err := json.MarshalIndent(dnsResult, "", "  ")
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("json encoding error: %v", err))
		return result, err
	}

	result.Data = append(result.Data, string(jsonData))
	return result, nil
}

func contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}
