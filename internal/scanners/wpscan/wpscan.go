package wpscan

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/IxBahy/ASM/internal/scanners"
	"github.com/IxBahy/ASM/pkg/client"
)

type WPScanScanner struct {
	config        scanners.ScannerConfig
	installState  scanners.InstallationState
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

	s := &WPScanScanner{
		config: config,
		installState: scanners.InstallationState{
			Installed: false,
			Version:   "",
		},
	}
	s.installState.Installed = s.IsInstalled()
	return s
}

// Setup ensures WPScan is installed
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

	return s.registerInstallationStats()
}

func (s *WPScanScanner) IsInstalled() bool {
	if !s.installState.Installed {
		if _, err := os.Stat(s.config.ExecutablePath); err == nil {
			s.registerInstallationStats()
		} else if _, err := exec.LookPath(s.config.Name); err == nil {
			s.registerInstallationStats()
		}
	}

	return s.installState.Installed
}

// Scan performs a WPScan scan on the specified target
func (s *WPScanScanner) Scan(target string) (scanners.ScannerResult, error) {
	if !s.IsInstalled() {
		return scanners.ScannerResult{}, fmt.Errorf("wpscan is not installed")
	}

	// Split the base command
	cmdParts := strings.Fields(s.config.Base_Command)

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

// GetConfig returns the scanner configuration
func (s *WPScanScanner) GetConfig() scanners.ScannerConfig {
	return s.config
}

// register marks the tool as installed and gets its version
func (s *WPScanScanner) registerInstallationStats() error {
	s.installState.Installed = true

	cmd := exec.Command(s.config.Name, "--version")
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
		// If version not found in output, use first line
		if s.installState.Version == "" && len(lines) > 0 {
			s.installState.Version = strings.TrimSpace(lines[0])
		}
	}

	fmt.Printf("WPScan registered as installed, version: %s\n", s.installState.Version)
	return nil
}
