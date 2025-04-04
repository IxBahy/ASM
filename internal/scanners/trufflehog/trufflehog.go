package trufflehog

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/IxBahy/ASM/internal/scanners"
	"github.com/IxBahy/ASM/pkg/client"
)

type TruffleHogScanner struct {
	*scanners.BaseScanner
	installClient client.ToolInstaller
}

func NewTruffleHogScanner() *TruffleHogScanner {
	config := scanners.ScannerConfig{
		Name:             "trufflehog",
		Version:          "latest",
		ExecutablePath:   "/usr/local/bin/trufflehog",
		Base_Command:     "trufflehog",
		InstallationType: client.InstallationTypeGithub,
		GithubOptions: scanners.GithubOptions{
			InstallLink:    "https://api.github.com/repos/trufflesecurity/trufflehog/releases/latest",
			InstallPattern: "trufflehog_.*_linux_amd64.tar.gz",
		},
	}

	base := &scanners.BaseScanner{
		Config: config,
		InstallState: scanners.InstallationState{
			Installed: false,
			Version:   "",
		},
	}

	s := &TruffleHogScanner{
		BaseScanner: base,
	}
	s.InstallState.Installed = s.IsInstalled()
	return s
}

func (s *TruffleHogScanner) Setup() error {
	if s.IsInstalled() {
		return nil
	}

	fmt.Println("Installing TruffleHog from GitHub releases...")

	installArgs := []string{
		s.Config.GithubOptions.InstallLink,
		s.Config.GithubOptions.InstallPattern,
		s.Config.Version,
		s.Config.ExecutablePath,
	}

	var err error
	s.installClient, err = client.ClientFactory(client.InstallationTypeGithub, installArgs, 5)
	if err != nil {
		return fmt.Errorf("failed to create GitHub client: %w", err)
	}

	if err := s.installClient.InstallTool(); err != nil {
		return fmt.Errorf("failed to install trufflehog: %w", err)
	}

	if path, err := exec.LookPath("trufflehog"); err == nil {
		s.Config.ExecutablePath = path
	}

	return s.RegisterInstallationStats()
}

func (s *TruffleHogScanner) Scan(target string) (scanners.ScannerResult, error) {
	if !s.IsInstalled() {
		return scanners.ScannerResult{}, fmt.Errorf("trufflehog is not installed")
	}

	result := scanners.ScannerResult{
		Data:   []string{},
		Errors: []string{},
	}

	isURL := strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://")

	cmdParts := strings.Fields(s.Config.Base_Command)

	if isURL {
		cmdParts = append(cmdParts, "git")
		cmdParts = append(cmdParts, "--json")
		cmdParts = append(cmdParts, "--no-update")
		cmdParts = append(cmdParts, target)
	} else {
		cmdParts = append(cmdParts, "filesystem")
		cmdParts = append(cmdParts, "--json")
		cmdParts = append(cmdParts, "--no-update")
		cmdParts = append(cmdParts, target)
	}

	cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
	output, err := cmd.CombinedOutput()

	for _, line := range strings.Split(string(output), "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			result.Data = append(result.Data, trimmed)
		}
	}

	if err != nil && len(result.Data) == 0 {
		result.Errors = append(result.Errors, fmt.Sprintf("command error: %v", err))
		result.Errors = append(result.Errors, string(output))
		return result, err
	}

	return result, nil
}
