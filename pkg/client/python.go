package client

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type PythonClient struct {
	packageName string
	version     string
	timeout     time.Duration
	pipArgs     []string
}

func NewPythonClient(install_args []string, timeout time.Duration) (*PythonClient, error) {
	if len(install_args) < 1 {
		return nil, fmt.Errorf("usage: python <package_name> [version] [extra_args...]")
	}

	packageName := install_args[0]

	version := "latest"
	if len(install_args) > 1 && install_args[1] != "" {
		version = install_args[1]
	}

	pipArgs := []string{"install"}

	if version != "latest" {
		pipArgs = append(pipArgs, fmt.Sprintf("%s==%s", packageName, version))
	} else {
		pipArgs = append(pipArgs, packageName)
	}

	if len(install_args) > 2 {
		pipArgs = append(pipArgs, install_args[2:]...)
	}

	return &PythonClient{
		packageName: packageName,
		version:     version,
		timeout:     timeout * time.Minute,
		pipArgs:     pipArgs,
	}, nil
}

func (c *PythonClient) InstallTool() error {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	pythonCmd := ""
	pipCmd := ""

	if _, err := exec.LookPath("python3"); err == nil {
		pythonCmd = "python"

		if _, err := exec.LookPath("pip"); err == nil {
			pipCmd = "pip"
		}
	}

	if pythonCmd == "" || pipCmd == "" {
		return fmt.Errorf("python and pip installation not found. Please install them ")
	}

	versionCmd := exec.CommandContext(ctx, pythonCmd, "--version")
	versionOutput, err := versionCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to get Python version: %w", err)
	}

	if versionOutput == nil {
		return fmt.Errorf("failed to get Python version: %w", err)
	} else if strings.Contains(string(versionOutput), "Python 2") {
		return fmt.Errorf("python 2 is not supported. Please install Python 3")
	}

	fmt.Printf("Using %s: %s\n", pythonCmd, strings.TrimSpace(string(versionOutput)))

	fmt.Printf("Installing %s %s using pip...\n", c.packageName, c.version)

	cmdArgs := append([]string{pipCmd}, c.pipArgs...)

	fmt.Printf("Running: %s\n", strings.Join(cmdArgs, " "))

	cmd := exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install %s with pip: %w", c.packageName, err)
	}

	_, err = exec.LookPath(c.packageName)
	if err == nil {
		fmt.Printf("Found %s executable in PATH\n", c.packageName)
		return nil
	}

	return nil
}
