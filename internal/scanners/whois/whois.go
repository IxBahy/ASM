package whois

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/IxBahy/ASM/internal/scanners"
	"github.com/IxBahy/ASM/pkg/client"
)

type WhoisScanner struct {
	*scanners.BaseScanner
	installClient client.ToolInstaller
}

func NewWhoisScanner() *WhoisScanner {
	config := scanners.ScannerConfig{
		Name:             "whois",
		Version:          "latest",
		ExecutablePath:   "/usr/bin/whois",
		Base_Command:     "whois",
		InstallationType: client.InstallationTypeShell,
	}

	base := &scanners.BaseScanner{
		Config: config,
		InstallState: scanners.InstallationState{
			Installed: false,
			Version:   "",
		},
	}
	s := &WhoisScanner{
		BaseScanner: base,
	}
	s.InstallState.Installed = s.IsInstalled()
	return s
}

func (s *WhoisScanner) Setup() error {
	if s.IsInstalled() {
		return nil
	}

	installArgs := []string{"whois", "-y"}
	var err error
	s.installClient, err = client.ClientFactory(s.Config.InstallationType, installArgs, 5)
	if err != nil {
		return fmt.Errorf("failed to create install client: %w", err)
	}

	if err := s.installClient.InstallTool(); err != nil {
		return fmt.Errorf("failed to install whois: %w", err)
	}

	return s.RegisterInstallationStats()
}

func (s *WhoisScanner) Scan(target string) (scanners.ScannerResult, error) {
	if !s.IsInstalled() {
		return scanners.ScannerResult{}, fmt.Errorf("whois is not installed")
	}

	re := regexp.MustCompile(`^(?:https?://)?(?:[^@\n]+@)?(?:www\.)?([^:/\n?]+)`)
	matches := re.FindStringSubmatch(target)
	if len(matches) > 1 {
		target = matches[1]
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
		result.Errors = append(result.Errors, fmt.Sprintf("scan warning: %v", err))
	}

	for _, line := range strings.Split(string(output), "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			if !strings.HasPrefix(trimmed, "%") && !strings.HasPrefix(trimmed, "#") {
				result.Data = append(result.Data, trimmed)
			}
		}
	}

	return result, nil
}
