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

func NewShellClient(install_args []string, timeout time.Duration) (*ShellClient, error) {
	if len(install_args) < 1 {

		return nil, fmt.Errorf("usage: shell <tool_name> [<args>]")
	}
	return &ShellClient{
		toolName: install_args[0],
		cmdArgs:  append([]string{"sudo", "apt", "install"}, install_args...),
		timeout:  timeout * time.Minute,
	}, nil
}

// InstallTool executes the commands specified in the ShellClient's cmdArgs to install
// the tool identified by toolName. It splits the commands on "&&" and executes them
// sequentially, displaying the output to stdout and stderr. The execution has a timeout
// defined by the client's timeout field. It returns an error if any command fails to execute
// or if the timeout is reached, otherwise it returns nil on successful installation.
func (c *ShellClient) InstallTool() error {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	fmt.Printf("Installing %s with command %s ...\n", c.toolName, c.cmdArgs)

	var cmdCommands [][]string
	currentCommand := []string{}

	for _, arg := range c.cmdArgs {
		if arg == "&&" {
			if len(currentCommand) > 0 {
				cmdCommands = append(cmdCommands, currentCommand)
				currentCommand = []string{}
			}
		} else {
			currentCommand = append(currentCommand, arg)
		}
	}

	if len(currentCommand) > 0 {
		cmdCommands = append(cmdCommands, currentCommand)
	}

	for _, cmdArgs := range cmdCommands {
		if len(cmdArgs) > 0 {
			cmd := exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to execute command '%v': %w", cmdArgs, err)
			}
		}
	}

	fmt.Printf("%s installed successfully\n", c.toolName)
	return nil
}
