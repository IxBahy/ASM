package nmap

import (
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/IxBahy/ASM/internal/scanners"
	"github.com/IxBahy/ASM/pkg/client"
)

type NmapScanner struct {
	*scanners.BaseScanner
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
	base := &scanners.BaseScanner{
		Config: config,
		InstallState: scanners.InstallationState{
			Installed: false,
			Version:   "",
		},
	}
	s := &NmapScanner{
		BaseScanner: base,
	}
	s.InstallState.Installed = s.IsInstalled()
	return s
}

func (s *NmapScanner) Setup() error {

	if s.IsInstalled() {
		return nil
	}

	installArgs := []string{"nmap", "-y"}
	var err error
	s.installClient, err = client.ClientFactory(s.Config.InstallationType, installArgs, 5)
	if err != nil {
		return fmt.Errorf("failed to create install client: %w", err)
	}

	if err := s.installClient.InstallTool(); err != nil {
		return fmt.Errorf("failed to install nmap: %w", err)
	}

	return s.RegisterInstallationStats()
}

func (s *NmapScanner) Scan(target string) (scanners.ScannerResult, error) {
	if !s.IsInstalled() {
		return scanners.ScannerResult{}, fmt.Errorf("nmap is not installed")
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
		result.Errors = append(result.Errors, fmt.Sprintf("scan error: %v", err))
		for _, line := range strings.Split(string(output), "\n") {
			if trimmed := strings.TrimSpace(line); trimmed != "" {
				result.Errors = append(result.Errors, trimmed)
			}
		}
		return result, err
	}

	for _, line := range strings.Split(string(output), "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			result.Data = append(result.Data, trimmed)
		}
	}

	return result, nil
}

type NmapRun struct {
	XMLName xml.Name `xml:"nmaprun"`
	Hosts   []Host   `xml:"host"`
}

type Host struct {
	Ports []Port `xml:"ports>port"`
}

type Port struct {
	PortID   string  `xml:"portid,attr"`
	Protocol string  `xml:"protocol,attr"`
	State    State   `xml:"state"`
	Service  Service `xml:"service"`
}

type State struct {
	State string `xml:"state,attr"`
}

type Service struct {
	Name    string `xml:"name,attr"`
	Product string `xml:"product,attr,omitempty"`
}

func (s *NmapScanner) ScanTopPorts(target string, top_count int) ([]Port, error) {
	cmdParts := strings.Fields(s.Config.Base_Command)
	tempFileName := "scanTopPorts.xml"
	cmdParts = append(cmdParts, "--top-ports", fmt.Sprintf("%d", top_count), target, "-oX", tempFileName)
	cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to run nmap scan: %w", err)
	}

	openPorts, err := filterOpenPortsInFile(tempFileName)
	if err != nil {
		return nil, err
	}
	if err := os.Remove(tempFileName); err != nil {
		fmt.Printf("Warning: Failed to remove temporary file %s: %v\n", tempFileName, err)
	}
	return openPorts, nil

}

func filterOpenPortsInFile(functionName string) ([]Port, error) {
	var nmapRun NmapRun
	xmlData, err := os.ReadFile(fmt.Sprintf("%s.xml", functionName))
	if err != nil {
		return nil, fmt.Errorf("failed to read XML file: %w", err)
	}

	xml.Unmarshal(xmlData, &nmapRun)

	openPorts := []Port{}
	for _, host := range nmapRun.Hosts {
		for _, port := range host.Ports {
			if port.State.State == "open" {
				openPorts = append(openPorts, port)
			}
		}
	}

	return openPorts, nil
}
