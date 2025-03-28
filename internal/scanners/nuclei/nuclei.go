package nuclei

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/IxBahy/ASM/internal/scanners"
	"github.com/IxBahy/ASM/pkg/client"
)

type NucleiScanner struct {
	config        scanners.ScannerConfig
	installState  scanners.InstallationState
	installClient client.ToolInstaller
}

// NewNucleiScanner creates a new Nuclei scanner
func NewNucleiScanner() *NucleiScanner {
	config := scanners.ScannerConfig{
		Name:    "nuclei",
		Version: "latest",
		GithubOptions: scanners.GithubOptions{
			InstallLink:    "https://api.github.com/repos/projectdiscovery/nuclei/releases/latest",
			InstallPattern: "nuclei_(.*)_linux_amd64.zip",
		},
		ExecutablePath:   "/usr/local/bin/nuclei",
		Base_Command:     "nuclei -t cves/ -u",
		InstallationType: client.InstallationTypeGithub,
	}
	s := &NucleiScanner{
		config: config,
		installState: scanners.InstallationState{
			Installed: false,
			Version:   "",
		},
	}
	s.installState.Installed = s.IsInstalled()
	return s
}

// Setup ensures Nuclei is installed
func (s *NucleiScanner) Setup() error {
	if s.IsInstalled() {
		return nil
	}

	installArgs := []string{
		s.config.GithubOptions.InstallLink,
		s.config.GithubOptions.InstallPattern,
		s.config.Version,
		s.config.ExecutablePath,
	}

	var err error
	s.installClient, err = client.ClientFactory(s.config.InstallationType, installArgs, 5)
	if err != nil {
		return fmt.Errorf("failed to create install client: %w", err)
	}

	if err := s.installClient.InstallTool(); err != nil {
		return fmt.Errorf("failed to install nuclei: %w", err)
	}

	return s.registerInstallationStats(s.config.Version)
}

func (s *NucleiScanner) IsInstalled() bool {
	if !s.installState.Installed {
		if _, err := os.Stat(s.config.ExecutablePath); err == nil {
			s.registerInstallationStats("")
		} else if _, err := exec.LookPath("nuclei"); err == nil {
			s.registerInstallationStats("")
		}
	}

	return s.installState.Installed
}

func (s *NucleiScanner) Scan(target string) (scanners.ScannerResult, error) {
	if !s.IsInstalled() {
		return scanners.ScannerResult{}, fmt.Errorf("nuclei is not installed")
	}

	cmdParts := strings.Fields(s.config.Base_Command)

	cmdParts = append(cmdParts, target)

	cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
	output, err := cmd.CombinedOutput()

	result := scanners.ScannerResult{
		Data:   []string{},
		Errors: []string{},
	}

	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("scan error: %v", err))
		for _, line := range strings.Split(string(output), "\n") {
			if trimmed := strings.TrimSpace(line); trimmed != "" {
				result.Errors = append(result.Errors, trimmed)
			}
		}
		return result, err
	}

	for _, line := range strings.Split(string(output), "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			result.Data = append(result.Data, trimmed)
		}
	}

	return result, nil
}

func (s *NucleiScanner) GetConfig() scanners.ScannerConfig {
	return s.config
}

func (s *NucleiScanner) registerInstallationStats(version string) error {
	s.installState.Installed = true

	if version != "" {
		s.installState.Version = version
	} else {
		cmd := exec.Command("nuclei", "-version")
		output, err := cmd.CombinedOutput()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.Contains(line, "Version:") {
					parts := strings.Split(line, ":")
					if len(parts) > 1 {
						s.installState.Version = strings.TrimSpace(parts[1])
						break
					}
				}
			}
		}
	}

	fmt.Printf("Nuclei registered as installed, version: %s\n", s.installState.Version)
	return nil
}
