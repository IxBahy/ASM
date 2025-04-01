package masscan

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/IxBahy/ASM/internal/scanners"
	"github.com/IxBahy/ASM/pkg/client"
	"github.com/IxBahy/ASM/pkg/utils/extractor"
)

type MassScanScanner struct {
	*scanners.BaseScanner
	installClient client.ToolInstaller
}

type MassScanResult struct {
	Host      string   `json:"host"`
	IP        string   `json:"ip"`
	Ports     []string `json:"ports"`
	TimeStamp string   `json:"timestamp"`
}

func NewMassScanScanner() *MassScanScanner {
	config := scanners.ScannerConfig{
		Name:             "masscan",
		Version:          "installed",
		ExecutablePath:   "/usr/bin/masscan",
		Base_Command:     "sudo masscan",
		InstallationType: client.InstallationTypeShell,
	}

	base := &scanners.BaseScanner{
		Config: config,
		InstallState: scanners.InstallationState{
			Installed: false,
			Version:   "",
		},
	}

	s := &MassScanScanner{
		BaseScanner: base,
	}
	s.InstallState.Installed = s.IsInstalled()
	return s
}

func (s *MassScanScanner) Setup() error {
	if s.IsInstalled() {
		return nil
	}

	installArgs := []string{"masscan", "libpcap-dev"}

	var err error
	s.installClient, err = client.ClientFactory(client.InstallationTypeShell, installArgs, 5)
	if err != nil {
		return fmt.Errorf("failed to create install client: %w", err)
	}

	if err := s.installClient.InstallTool(); err != nil {
		return fmt.Errorf("failed to install masscan: %w", err)
	}

	return s.RegisterInstallationStats()
}

func (s *MassScanScanner) IsInstalled() bool {
	if !s.InstallState.Installed {
		if _, err := os.Stat(s.Config.ExecutablePath); err == nil {
			s.RegisterInstallationStats()
		} else if _, err := exec.LookPath("masscan"); err == nil {
			s.Config.ExecutablePath, _ = exec.LookPath("masscan")
			s.RegisterInstallationStats()
		}
	}
	return s.InstallState.Installed
}

func (s *MassScanScanner) Scan(target string) (scanners.ScannerResult, error) {
	if !s.IsInstalled() {
		return scanners.ScannerResult{}, fmt.Errorf("masscan is not installed")
	}

	result := scanners.ScannerResult{
		Data:   []string{},
		Errors: []string{},
	}

	targetIP, err := resolveTarget(target)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("failed to resolve target: %v", err))
		return result, err
	}

	outputFile, err := os.CreateTemp("", "masscan-*.json")
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("failed to create temp file: %v", err))
		return result, err
	}
	defer os.Remove(outputFile.Name())
	outputFile.Close()

	cmdArgs := []string{
		"sudo",
		"masscan",
		"-p80,443,8000-8100",
		targetIP,
		"--rate=1000",
		"--wait=0",
		"-oJ", outputFile.Name(),
	}

	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmdOutput, err := cmd.CombinedOutput()

	if err != nil {

		if strings.Contains(string(cmdOutput), "caught signal") || !strings.Contains(string(cmdOutput), "error") {

		} else {
			result.Errors = append(result.Errors, fmt.Sprintf("masscan error: %v", err))
			result.Errors = append(result.Errors, string(cmdOutput))

		}
	}

	jsonData, err := os.ReadFile(outputFile.Name())
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("failed to read output file: %v", err))

		jsonData = []byte("[]")
	}

	var portList []string

	if len(jsonData) > 2 {

		jsonStr := string(jsonData)
		jsonStr = strings.Trim(jsonStr, "[]\n\r ")

		if jsonStr != "" {

			for _, line := range strings.Split(jsonStr, ",\n") {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}

				if portStart := strings.Index(line, "\"port\": "); portStart > 0 {
					portEnd := strings.Index(line[portStart+8:], ",")
					if portEnd > 0 {
						portStr := line[portStart+8 : portStart+8+portEnd]
						portList = append(portList, portStr+"/tcp")
					}
				}
			}
		}
	}

	scanResult := MassScanResult{
		Host:      target,
		IP:        targetIP,
		Ports:     portList,
		TimeStamp: time.Now().Format(time.RFC3339),
	}

	resultJSON, err := json.MarshalIndent(scanResult, "", "  ")
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("failed to encode result: %v", err))
		return result, err
	}

	result.Data = append(result.Data, string(resultJSON))
	return result, nil
}

func (s *MassScanScanner) RegisterInstallationStats() error {
	s.InstallState.Installed = true

	cmd := exec.Command("masscan", "--version")
	output, err := cmd.CombinedOutput()

	if err != nil {
		s.InstallState.Version = "unknown"
	} else {
		versionOutput := strings.TrimSpace(string(output))
		if versionOutput != "" {

			if idx := strings.Index(versionOutput, "\n"); idx > 0 {
				versionOutput = versionOutput[:idx]
			}
			s.InstallState.Version = versionOutput
		} else {
			s.InstallState.Version = "installed"
		}
	}

	fmt.Printf("Masscan registered as installed, version: %s\n", s.InstallState.Version)
	return nil
}

func resolveTarget(target string) (string, error) {

	host := extractor.ExtractDomain(target)

	if net.ParseIP(host) != nil {
		return host, nil
	}

	ips, err := net.LookupHost(host)
	if err != nil {
		return "", fmt.Errorf("failed to resolve host: %w", err)
	}

	if len(ips) == 0 {
		return "", fmt.Errorf("no IP addresses found for host")
	}

	return ips[0], nil
}
