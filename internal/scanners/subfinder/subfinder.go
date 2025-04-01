package subfinder

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

type SubfinderScanner struct {
	*scanners.BaseScanner
	installClient client.ToolInstaller
}

type SubdomainResult struct {
	Domain     string   `json:"domain"`
	Subdomains []string `json:"subdomains"`
	TimeStamp  string   `json:"timestamp"`
}

func NewSubfinderScanner() *SubfinderScanner {
	config := scanners.ScannerConfig{
		Name:             "subfinder",
		Version:          "latest",
		ExecutablePath:   "/usr/local/bin/subfinder",
		Base_Command:     "subfinder -d",
		InstallationType: client.InstallationTypeGithub,
		GithubOptions: scanners.GithubOptions{
			InstallLink:    "https://api.github.com/repos/projectdiscovery/subfinder/releases/latest",
			InstallPattern: "subfinder_.*_linux_amd64.zip",
		},
	}

	base := &scanners.BaseScanner{
		Config: config,
		InstallState: scanners.InstallationState{
			Installed: false,
			Version:   "",
		},
	}

	s := &SubfinderScanner{
		BaseScanner: base,
	}
	s.InstallState.Installed = s.IsInstalled()
	return s
}

func (s *SubfinderScanner) Setup() error {
	if s.IsInstalled() {
		return nil
	}

	// Install from GitHub release
	installArgs := []string{
		s.Config.GithubOptions.InstallLink,
		s.Config.GithubOptions.InstallPattern,
		s.Config.Version,
		s.Config.ExecutablePath,
	}

	var err error
	s.installClient, err = client.ClientFactory(client.InstallationTypeGithub, installArgs, 5)
	if err != nil {
		return fmt.Errorf("failed to create install client: %w", err)
	}

	if err := s.installClient.InstallTool(); err != nil {
		return fmt.Errorf("failed to install subfinder: %w", err)
	}

	return s.RegisterInstallationStats()
}

func (s *SubfinderScanner) Scan(domain string) (scanners.ScannerResult, error) {
	if !s.IsInstalled() {
		return scanners.ScannerResult{}, fmt.Errorf("subfinder is not installed")
	}

	result := scanners.ScannerResult{
		Data:   []string{},
		Errors: []string{},
	}

	domain = extractor.ExtractDomain(domain)

	cmdParts := strings.Fields(s.Config.Base_Command)
	cmdParts = append(cmdParts, domain, "-silent", "-json")

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
			Host string `json:"host"`
		}

		if err := json.Unmarshal([]byte(line), &entry); err == nil && entry.Host != "" {
			subdomains = append(subdomains, entry.Host)
		}
	}

	subdomainResult := SubdomainResult{
		Domain:     domain,
		Subdomains: subdomains,
		TimeStamp:  time.Now().Format(time.RFC3339),
	}

	resultJSON, err := json.MarshalIndent(subdomainResult, "", "  ")
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("json encoding error: %v", err))
		return result, err
	}

	result.Data = append(result.Data, string(resultJSON))
	return result, nil
}

func (s *SubfinderScanner) RegisterInstallationStats() error {
	s.InstallState.Installed = true

	cmd := exec.Command(s.Config.ExecutablePath, "-version")
	output, err := cmd.CombinedOutput()

	if err != nil {
		s.InstallState.Version = "unknown"
	} else {
		versionOutput := strings.TrimSpace(string(output))
		if versionOutput != "" {
			s.InstallState.Version = versionOutput
		} else {
			s.InstallState.Version = "installed"
		}
	}

	fmt.Printf("Subfinder registered as installed, version: %s\n", s.InstallState.Version)
	return nil
}
