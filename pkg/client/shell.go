package client

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
)

type ShellClient struct {
	timeout  time.Duration
	cmdArgs  []string
	toolName string
}

func NewShellClient(install_args []string, timeout time.Duration) *ShellClient {

	return &ShellClient{
		toolName: install_args[0],
		cmdArgs:  append([]string{"sudo", "apt", "install"}, install_args...),
		timeout:  timeout * time.Minute,
	}
}

func (c *ShellClient) InstallTool() error {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	fmt.Printf("Installing %s with command %s ...\n", c.toolName, c.cmdArgs)

	cmd := exec.CommandContext(ctx, c.cmdArgs[0], c.cmdArgs[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install %s: %w", c.toolName, err)
	}

	fmt.Printf("%s installed successfully\n", c.toolName)
	return nil
}
