package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/IxBahy/ASM/pkg/client/utils"
)

type GithubClient struct {
	httpClient  *http.Client
	version     string
	DownloadUrl string
	destPath    string
}

func NewGithubClient(install_args []string, timeout time.Duration) (*GithubClient, error) {
	if len(install_args) != 3 {
		return nil, fmt.Errorf("usage: github <url> <version> <dest_path>")
	}
	url := install_args[0]
	version := install_args[1]
	destPath := install_args[2]
	httpClient := &http.Client{
		Timeout: timeout * time.Minute,
	}
	return &GithubClient{
		DownloadUrl: url,
		version:     version,
		destPath:    destPath,
		httpClient:  httpClient,
	}, nil
}
func (c *GithubClient) InstallTool() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	toolName := filepath.Base(c.destPath)

	installDir := filepath.Dir(c.destPath)
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", installDir, err)
	}

	tempFile, err := os.CreateTemp("", "tool-download-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	fmt.Printf("Downloading %s, version :: %s ...\n", toolName, c.version)
	if err := c.downloadFile(ctx, c.DownloadUrl, tempFile); err != nil {
		return fmt.Errorf("failed to download %s: %w", toolName, err)
	}

	if utils.IsArchive(c.DownloadUrl) {
		if err := utils.ExtractExecutable(tempFile.Name(), c.destPath, toolName); err != nil {
			return fmt.Errorf("failed to extract executable: %w", err)
		}
	} else {
		if err := utils.CopyFile(tempFile.Name(), c.destPath); err != nil {
			return fmt.Errorf("failed to copy executable: %w", err)
		}
	}

	if err := os.Chmod(c.destPath, 0755); err != nil {
		return fmt.Errorf("failed to make %s executable: %w", c.destPath, err)
	}

	fmt.Printf("%s %s installed successfully at %s\n", toolName, c.version, c.destPath)
	return nil
}

func (c *GithubClient) downloadFile(ctx context.Context, url string, w io.Writer) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	_, err = io.Copy(w, resp.Body)
	return err
}
