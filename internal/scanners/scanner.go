package scanners

import (
	"github.com/IxBahy/ASM/pkg/client"
	"github.com/IxBahy/ASM/pkg/interfaces"
)

type Scanner interface {
	interfaces.Installable
	Scan(target string) (ScannerResult, error)
	GetConfig() ScannerConfig
}

type ScannerConfig struct {
	Name             string
	Version          string
	InstallLink      string
	ExecutablePath   string
	Base_Command     string
	LocalPath        string
	InstallationType client.InstallationType
}

type ScannerResult struct {
	Data   []string
	Errors []string
}
