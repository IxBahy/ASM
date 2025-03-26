package clients

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
)

type ShellClient struct {
	timeout time.Duration
}

func NewShellClient(timeout time.Duration) *ShellClient {
	return &ShellClient{
		timeout: timeout * time.Minute,
	}
}

func (c *ShellClient) Install(cmdArgs []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	toolName := "package"
	if len(cmdArgs) > 0 {
		toolName = cmdArgs[0]
	}

	fmt.Printf("Installing %s with command %s ...\n", toolName, cmdArgs)

	cmdBase := []string{"sudo", "apt", "install"}

	fullCmd := append(cmdBase, cmdArgs...)

	cmd := exec.CommandContext(ctx, fullCmd[0], fullCmd[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install %s: %w", toolName, err)
	}

	fmt.Printf("%s installed successfully\n", toolName)
	return nil
}
