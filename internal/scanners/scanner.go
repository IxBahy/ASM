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
