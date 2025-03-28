package scanners

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Scanner interface {
	Setup() error
	Scan(target string) (ScannerResult, error)
}
type BaseScanner struct {
	config       ScannerConfig
	installState InstallationState
}
type InstallationState struct {
	Installed bool
	Version   string
	Path      string
}
type InstallationType string

const (
	InstallationTypeGithub InstallationType = "github"
	InstallationTypeShell  InstallationType = "shell"
)

type ScannerConfig struct {
	Name             string
	Version          string
	InstallLink      string
	ExecutablePath   string
	Base_Command     string
	LocalPath        string
	InstallationType InstallationType
}

type ScannerResult struct {
	Data   []string
	Errors []string
}

func NewBaseScanner(config ScannerConfig, installationType InstallationType) *BaseScanner {
	return &BaseScanner{
		config:           config,
		InstallationType: installationType,
		installState: InstallationState{
			Installed: false,
		},
	}
}

// Config returns the scanner configuration
func (s *BaseScanner) Config() ScannerConfig {
	return s.config
}

// IsInstalled returns true if the scanner is installed
func (s *BaseScanner) IsInstalled() bool {
	return s.installState.Installed
}

// Setup ensures the scanner tool is installed
func (s *BaseScanner) Setup() error {
	// Check if tool is already installed
	if s.isToolAvailable() {
		s.registerInstallation(s.config.Version, s.config.ExecutablePath)
		return nil
	}

	// Create installation directory if needed
	installDir := filepath.Dir(s.config.ExecutablePath)
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", installDir, err)
	}

	var err error
	if s.useShellSetup {
		err = s.setupWithShellClient()
	} else {
		err = s.setupWithGithubClient()
	}

	if err != nil {
		return err
	}

	// Verify installation
	if !s.isToolAvailable() {
		return fmt.Errorf("tool %s was not installed correctly", s.config.Name)
	}

	s.registerInstallation(s.config.Version, s.config.ExecutablePath)
	return nil
}

// setupWithGithubClient installs the tool using the GitHub client
func (s *BaseScanner) setupWithGithubClient() error {
	if s.githubClient == nil {
		return fmt.Errorf("GitHub client not available")
	}

	fmt.Printf("Installing %s using GitHub client...\n", s.config.Name)
	return s.githubClient.InstallTool(s.config.InstallLink, s.config.Version, s.config.ExecutablePath)
}

// setupWithShellClient installs the tool using the Shell client
func (s *BaseScanner) setupWithShellClient() error {
	if s.shellClient == nil {
		return fmt.Errorf("shell client not available")
	}

	fmt.Printf("Installing %s using Shell client...\n", s.config.Name)

	// Parse the tool name and create command arguments
	toolName := s.config.Name
	cmdArgs := []string{toolName}

	// Add version if specified
	if s.config.Version != "" && s.config.Version != "latest" {
		cmdArgs = append(cmdArgs, s.config.Version)
	}

	// Add additional flags
	cmdArgs = append(cmdArgs, "-y")

	return s.shellClient.Install(cmdArgs)
}

// Scan executes the scanner with the provided target
func (s *BaseScanner) Scan(target string) (ScannerResult, error) {
	if !s.installState.Installed {
		return ScannerResult{}, fmt.Errorf("scanner %s is not installed", s.config.Name)
	}

	// Prepare command parts
	cmdParts := strings.Fields(s.config.Base_Command)
	if len(cmdParts) == 0 {
		cmdParts = []string{s.config.ExecutablePath}
	}

	// Add target
	cmdParts = append(cmdParts, target)

	// Execute command
	cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
	output, err := cmd.CombinedOutput()

	result := ScannerResult{
		Data:   []string{},
		Errors: []string{},
	}

	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("scan error: %v", err))
		result.Errors = append(result.Errors, string(output))
		return result, err
	}

	// Process output into result.Data
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line = strings.TrimSpace(line); line != "" {
			result.Data = append(result.Data, line)
		}
	}

	return result, nil
}

// registerInstallation updates the installation state
func (s *BaseScanner) registerInstallation(version, path string) {
	s.installState.Installed = true
	s.installState.Version = version
	s.installState.Path = path
	fmt.Printf("Registered %s version %s at %s\n", s.config.Name, version, path)
}

// isToolAvailable checks if the tool is available in the system
func (s *BaseScanner) isToolAvailable() bool {
	// Check local executable path
	if _, err := os.Stat(s.config.ExecutablePath); err == nil {
		return true
	}

	// Check if tool is in PATH
	_, err := exec.LookPath(s.config.Name)
	return err == nil
}
