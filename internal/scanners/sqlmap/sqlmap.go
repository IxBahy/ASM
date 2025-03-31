package sqlmap

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/IxBahy/ASM/internal/scanners"
	"github.com/IxBahy/ASM/pkg/client"
)

type SQLMapScanner struct {
	*scanners.BaseScanner
	installClient client.ToolInstaller
}

func NewSQLMapScanner() *SQLMapScanner {
	config := scanners.ScannerConfig{
		Name:             "sqlmap",
		Version:          "latest",
		ExecutablePath:   "/usr/local/bin/sqlmap",
		Base_Command:     "sqlmap -u",
		InstallationType: client.InstallationTypePython,
	}

	base := &scanners.BaseScanner{
		Config: config,
		InstallState: scanners.InstallationState{
			Installed: false,
			Version:   "",
		},
	}

	s := &SQLMapScanner{
		BaseScanner: base,
	}
	s.InstallState.Installed = s.IsInstalled()
	return s
}

func (s *SQLMapScanner) Setup() error {
	if s.IsInstalled() {
		return nil
	}

	installArgs := []string{"sqlmap", s.Config.Version}

	var err error
	s.installClient, err = client.ClientFactory(s.Config.InstallationType, installArgs, 5)
	if err != nil {
		return fmt.Errorf("failed to create install client: %w", err)
	}

	if err := s.installClient.InstallTool(); err != nil {
		return fmt.Errorf("failed to install sqlmap: %w", err)
	}

	sqlmapPath, lookErr := exec.LookPath("sqlmap")
	if lookErr == nil {
		s.Config.ExecutablePath = sqlmapPath
	}

	return s.RegisterInstallationStats()
}

func (s *SQLMapScanner) Scan(target string) (scanners.ScannerResult, error) {
	if !s.IsInstalled() {
		return scanners.ScannerResult{}, fmt.Errorf("sqlmap is not installed")
	}

	tmpDir, err := os.MkdirTemp("", "sqlmap-output-*")
	if err != nil {
		return scanners.ScannerResult{}, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	cmdParts := strings.Fields(s.Config.Base_Command)
	cmdParts = append(cmdParts, target)

	cmdParts = append(cmdParts,
		"--batch",
		"--output-dir", tmpDir,
		"--forms",
		"--level", "1",
		"--risk", "1",
		"--timeout", "30",
		"--json-output", filepath.Join(tmpDir, "results.json"),
	)

	cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
	output, err := cmd.CombinedOutput()

	result := scanners.ScannerResult{
		Data:   []string{},
		Errors: []string{},
	}

	for _, line := range strings.Split(string(output), "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			result.Data = append(result.Data, trimmed)
		}
	}

	jsonFile := filepath.Join(tmpDir, "results.json")
	if jsonData, readErr := os.ReadFile(jsonFile); readErr == nil {
		result.Data = append(result.Data, string(jsonData))
	}

	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("scan error: %v", err))
		return result, err
	}

	return result, nil
}

func (s *SQLMapScanner) RegisterInstallationStats() error {
	s.InstallState.Installed = true

	var cmd *exec.Cmd

	if _, err := os.Stat(s.Config.ExecutablePath); err == nil {
		cmd = exec.Command(s.Config.ExecutablePath, "--version")
	} else {
		sqlmapPath, err := exec.LookPath("sqlmap")
		if err != nil {
			s.InstallState.Version = "unknown"
			return nil
		}
		cmd = exec.Command(sqlmapPath, "--version")
		s.Config.ExecutablePath = sqlmapPath
	}

	output, err := cmd.CombinedOutput()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "Version") {
				parts := strings.Fields(line)
				if len(parts) > 0 {
					s.InstallState.Version = parts[len(parts)-1]
					break
				}
			}
		}
	}

	fmt.Printf("SQLMap registered as installed, version: %s\n", s.InstallState.Version)
	return nil
}
