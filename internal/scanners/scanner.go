package scanners

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/IxBahy/ASM/pkg/client"
	"github.com/IxBahy/ASM/pkg/interfaces"
)

type Scanner interface {
	interfaces.Installable
	Scan(target string) (ScannerResult, error)
	GetConfig() ScannerConfig
	GetInstallationState() InstallationState
	IsInstalled() bool
	RegisterInstallationStats() error
}

type ScannerConfig struct {
	Name             string
	Version          string
	GithubOptions    GithubOptions
	ExecutablePath   string
	Base_Command     string
	InstallationType client.InstallationType
}
type GithubOptions struct {
	InstallLink    string
	InstallPattern string
}
type ScannerResult struct {
	Data   []string
	Errors []string
}
type InstallationState struct {
	Installed bool
	Version   string
}

type BaseScanner struct {
	Config       ScannerConfig
	InstallState InstallationState
}

func (s *BaseScanner) IsInstalled() bool {
	if !s.InstallState.Installed {
		if _, err := os.Stat(s.Config.ExecutablePath); err == nil {
			s.RegisterInstallationStats()
		} else if _, err := exec.LookPath(s.Config.Name); err == nil {
			s.RegisterInstallationStats()
		}
	}

	return s.InstallState.Installed
}
func (b *BaseScanner) GetConfig() ScannerConfig {
	return b.Config
}
func (s *BaseScanner) RegisterInstallationStats() error {
	s.InstallState.Installed = true

	cmd := exec.Command(s.Config.Name, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error executing command: %s\n", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if regexp.MustCompile(`(?i)version`).MatchString(line) {
			s.InstallState.Version = strings.TrimSpace(line)
			break
		}
	}
	if s.InstallState.Version == "" && len(lines) > 0 && err == nil {
		s.InstallState.Version = strings.TrimSpace(string(output))
	}

	fmt.Printf("WPScan registered as installed, version: %s\n", s.InstallState.Version)
	return nil
}

func (s *BaseScanner) GetInstallationState() InstallationState {
	return s.InstallState
}
