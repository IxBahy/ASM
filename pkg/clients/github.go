package clients

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/IxBahy/ASM/pkg/clients/client_utils"
)

type GithubClient struct {
	httpClient *http.Client
}

func NewGithubClient(timeout time.Duration) *GithubClient {
	return &GithubClient{
		httpClient: &http.Client{
			Timeout: timeout * time.Minute,
		},
	}
}

func (c *GithubClient) InstallTool(url, version, destPath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	toolName := filepath.Base(destPath)

	installDir := filepath.Dir(destPath)
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", installDir, err)
	}

	tempFile, err := os.CreateTemp("", "tool-download-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Download the file
	fmt.Printf("Downloading %s, version :: %s ...\n", toolName, version)
	if err := c.downloadFile(ctx, url, tempFile); err != nil {
		return fmt.Errorf("failed to download %s: %w", toolName, err)
	}

	if client_utils.IsArchive(url) {
		if err := client_utils.ExtractExecutable(tempFile.Name(), destPath, toolName); err != nil {
			return fmt.Errorf("failed to extract executable: %w", err)
		}
	} else {
		if err := client_utils.CopyFile(tempFile.Name(), destPath); err != nil {
			return fmt.Errorf("failed to copy executable: %w", err)
		}
	}

	if err := os.Chmod(destPath, 0755); err != nil {
		return fmt.Errorf("failed to make %s executable: %w", destPath, err)
	}

	fmt.Printf("%s %s installed successfully at %s\n", toolName, version, destPath)
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
