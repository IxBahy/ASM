package semgrep

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/IxBahy/ASM/internal/scanners"
	"github.com/IxBahy/ASM/pkg/client"
)

type SemgrepScanner struct {
	*scanners.BaseScanner
	installClient client.ToolInstaller
}

func NewSemgrepScanner() *SemgrepScanner {
	config := scanners.ScannerConfig{
		Name:             "semgrep",
		Version:          "latest",
		ExecutablePath:   "/usr/local/bin/semgrep",
		Base_Command:     "semgrep scan",
		InstallationType: client.InstallationTypePython,
	}

	base := &scanners.BaseScanner{
		Config: config,
		InstallState: scanners.InstallationState{
			Installed: false,
			Version:   "",
		},
	}

	s := &SemgrepScanner{
		BaseScanner: base,
	}
	s.InstallState.Installed = s.IsInstalled()
	return s
}

func (s *SemgrepScanner) Setup() error {
	if s.IsInstalled() {
		return nil
	}

	installArgs := []string{"semgrep", s.Config.Version}

	var err error
	s.installClient, err = client.ClientFactory(client.InstallationTypePython, installArgs, 5)
	if err != nil {
		return fmt.Errorf("failed to create install client: %w", err)
	}

	if err := s.installClient.InstallTool(); err != nil {
		return fmt.Errorf("failed to install semgrep: %w", err)
	}

	if path, err := exec.LookPath("semgrep"); err == nil {
		s.Config.ExecutablePath = path
	}

	return s.RegisterInstallationStats()
}

func (s *SemgrepScanner) IsInstalled() bool {
	if !s.InstallState.Installed {
		if _, err := os.Stat(s.Config.ExecutablePath); err == nil {
			s.RegisterInstallationStats()
		} else if _, err := exec.LookPath("semgrep"); err == nil {
			s.Config.ExecutablePath, _ = exec.LookPath("semgrep")
			s.RegisterInstallationStats()
		}
	}

	return s.InstallState.Installed
}

func (s *SemgrepScanner) Scan(target string) (scanners.ScannerResult, error) {
	if !s.IsInstalled() {
		return scanners.ScannerResult{}, fmt.Errorf("semgrep is not installed")
	}

	result := scanners.ScannerResult{
		Data:   []string{},
		Errors: []string{},
	}

	fileInfo, err := os.Stat(target)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("error accessing target: %v", err))
		return result, err
	}

	cmdParts := strings.Fields(s.Config.Base_Command)

	cmdParts = append(cmdParts, "--json")

	cmdParts = append(cmdParts, "--config=auto")

	cmdParts = append(cmdParts, target)

	if !fileInfo.IsDir() {

	} else {

		cmdParts = append(cmdParts, "--exclude=node_modules,dist,build,vendor")
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
		return result, err
	}

	return result, nil
}

func (s *SemgrepScanner) RegisterInstallationStats() error {
	s.InstallState.Installed = true

	cmd := exec.Command(s.Config.ExecutablePath, "--version")
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

	fmt.Printf("Semgrep registered as installed, version: %s\n", s.InstallState.Version)
	return nil
}
