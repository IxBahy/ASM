package nuclei

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/IxBahy/ASM/internal/scanners"
	"github.com/IxBahy/ASM/pkg/client"
)

type NucleiScanner struct {
	*scanners.BaseScanner
	installClient client.ToolInstaller
}

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
	base := &scanners.BaseScanner{
		Config: config,
		InstallState: scanners.InstallationState{
			Installed: false,
			Version:   "",
		},
	}
	s := &NucleiScanner{
		BaseScanner: base,
	}
	s.InstallState.Installed = s.IsInstalled()
	return s
}

func (s *NucleiScanner) Setup() error {
	if s.IsInstalled() {
		return nil
	}

	installArgs := []string{
		s.Config.GithubOptions.InstallLink,
		s.Config.GithubOptions.InstallPattern,
		s.Config.Version,
		s.Config.ExecutablePath,
	}

	var err error
	s.installClient, err = client.ClientFactory(s.Config.InstallationType, installArgs, 5)
	if err != nil {
		return fmt.Errorf("failed to create install client: %w", err)
	}

	if err := s.installClient.InstallTool(); err != nil {
		return fmt.Errorf("failed to install nuclei: %w", err)
	}

	return s.RegisterInstallationStats()
}

func (s *NucleiScanner) Scan(target string) (scanners.ScannerResult, error) {
	if !s.IsInstalled() {
		return scanners.ScannerResult{}, fmt.Errorf("nuclei is not installed")
	}

	cmdParts := strings.Fields(s.Config.Base_Command)

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
