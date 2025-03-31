package client

import (
	"fmt"
	"time"
)

type InstallationType string

const (
	InstallationTypeGithub   InstallationType = "github"
	InstallationTypeShell    InstallationType = "shell"
	InstallationTypePython   InstallationType = "python"
	InstallationTypeInternal InstallationType = "internal"
)

func ClientFactory(installType InstallationType, install_args []string, timeout time.Duration) (ToolInstaller, error) {
	var client ToolInstaller
	var err error

	switch installType {
	case InstallationTypeGithub:
		client, err = NewGithubClient(install_args, timeout)

	case InstallationTypeShell:
		client, err = NewShellClient(install_args, timeout)

	case InstallationTypeInternal:
		// TODO: Implement internal installation logic
		// For now, we return an error
		client = nil
		err = fmt.Errorf("internal installation type is not implemented yet")

	case InstallationTypePython:
		client, err = NewPythonClient(install_args, timeout)

	default:
		client = nil
		err = fmt.Errorf("unsupported installation type: %s", installType)
	}

	return client, err
}
