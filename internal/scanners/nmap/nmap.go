package nmap

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/IxBahy/ASM/internal/scanners"
	"github.com/IxBahy/ASM/pkg/client"
)

type NmapScanner struct {
	config        scanners.ScannerConfig
	installState  scanners.InstallationState
	installClient client.ToolInstaller
}

func NewNmapScanner() *NmapScanner {
	config := scanners.ScannerConfig{
		Name:             "nmap",
		Version:          "latest",
		ExecutablePath:   "/usr/bin/nmap",
		Base_Command:     "nmap -sV -T4",
		InstallationType: client.InstallationTypeShell,
	}

	s := &NmapScanner{
		config: config,
		installState: scanners.InstallationState{
			Installed: false,
			Version:   "",
		},
	}
	s.installState.Installed = s.IsInstalled()
	return s
}

func (s *NmapScanner) Setup() error {

	if s.IsInstalled() {
		return nil
	}

	installArgs := []string{"nmap", "-y"}
	var err error
	s.installClient, err = client.ClientFactory(s.config.InstallationType, installArgs, 5)
	if err != nil {
		return fmt.Errorf("failed to create install client: %w", err)
	}

	if err := s.installClient.InstallTool(); err != nil {
		return fmt.Errorf("failed to install nmap: %w", err)
	}

	return s.registerInstallationStats()
}

func (s *NmapScanner) IsInstalled() bool {

	if !s.installState.Installed {
		if _, err := os.Stat(s.config.ExecutablePath); err == nil {
			s.registerInstallationStats()
		} else if _, err := exec.LookPath("nmap"); err == nil {
			s.registerInstallationStats()
		}
	}

	return s.installState.Installed
}

// GetConfig returns the scanner configuration
func (s *NmapScanner) GetConfig() scanners.ScannerConfig {
	return s.config
}

// register marks the tool as installed and gets its version
func (s *NmapScanner) registerInstallationStats() error {
	s.installState.Installed = true

	cmd := exec.Command("nmap", "--version")
	output, err := cmd.CombinedOutput()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		if len(lines) > 0 {
			versionInfo := strings.TrimSpace(lines[0])
			s.installState.Version = versionInfo
		}
	}

	fmt.Printf("Nmap registered as installed, version: %s\n", s.installState.Version)
	return nil
}

func (s *NmapScanner) Scan(target string) (scanners.ScannerResult, error) {
	if !s.IsInstalled() {
		return scanners.ScannerResult{}, fmt.Errorf("nmap is not installed")
	}

	// Split the base command
	cmdParts := strings.Fields(s.config.Base_Command)

	// Add the target
	cmdParts = append(cmdParts, target)

	// Execute the command
	cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
	output, err := cmd.CombinedOutput()

	result := scanners.ScannerResult{
		Data:   []string{},
		Errors: []string{},
	}

	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("scan error: %v", err))
		// Include output in errors even if there was an error
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
