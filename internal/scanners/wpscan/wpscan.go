package wpscan

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/IxBahy/ASM/internal/scanners"
	"github.com/IxBahy/ASM/pkg/client"
)

type WPScanScanner struct {
	*scanners.BaseScanner
	installClient client.ToolInstaller
}

func NewWPScanScanner() *WPScanScanner {
	config := scanners.ScannerConfig{
		Name:             "wpscan",
		Version:          "latest",
		ExecutablePath:   "/usr/local/bin/wpscan",
		Base_Command:     "wpscan --url",
		InstallationType: client.InstallationTypeShell,
	}

	base := &scanners.BaseScanner{
		Config: config,
		InstallState: scanners.InstallationState{
			Installed: false,
			Version:   "",
		},
	}
	s := &WPScanScanner{
		BaseScanner: base,
	}
	s.InstallState.Installed = s.IsInstalled()
	return s
}

func (s *WPScanScanner) Setup() error {
	if s.IsInstalled() {
		return nil
	}

	depsArgs := []string{
		"ruby",
		"ruby-dev",
		"git",
		"curl",
		"libcurl4-openssl-dev",
		"make",
		"zlib1g-dev",
		"gawk",
		"g++",
		"gcc",
		"-y",
		"&&",
		"sudo",
		"gem",
		"install",
		"wpscan",
	}
	var err error

	s.installClient, err = client.ClientFactory(client.InstallationTypeShell, depsArgs, 5)
	if err != nil {
		return fmt.Errorf("failed to create install client for dependencies: %w", err)
	}

	fmt.Println("Installing dependencies for WPScan...")
	if err := s.installClient.InstallTool(); err != nil {
		return fmt.Errorf("failed to install wpscan with it dependencies: %w", err)
	}

	return s.RegisterInstallationStats()
}

// Scan performs a WPScan scan on the specified target
func (s *WPScanScanner) Scan(target string) (scanners.ScannerResult, error) {
	if !s.IsInstalled() {
		return scanners.ScannerResult{}, fmt.Errorf("wpscan is not installed")
	}

	// Split the base command
	cmdParts := strings.Fields(s.Config.Base_Command)

	// Add the target
	cmdParts = append(cmdParts, target)

	// Add format as json for better parsing
	cmdParts = append(cmdParts, "--format", "json")

	// Execute the command
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

	// Process successful output
	for _, line := range strings.Split(string(output), "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			result.Data = append(result.Data, trimmed)
		}
	}

	return result, nil
}
