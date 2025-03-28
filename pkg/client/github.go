package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/IxBahy/ASM/pkg/client/utils"
)

type GithubRelease struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type GithubClient struct {
	httpClient   *http.Client
	version      string
	DownloadUrl  string
	destPath     string
	assetPattern string
}

func NewGithubClient(install_args []string, timeout time.Duration) (*GithubClient, error) {
	if len(install_args) < 4 {
		return nil, fmt.Errorf("usage: github <url> <version> <dest_path> <asset_pattern>")
	}

	url := install_args[0]
	assetPattern := install_args[1]
	version := install_args[2]
	destPath := install_args[3]

	httpClient := &http.Client{
		Timeout: timeout * time.Minute,
	}

	return &GithubClient{
		DownloadUrl:  url,
		version:      version,
		destPath:     destPath,
		httpClient:   httpClient,
		assetPattern: assetPattern,
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

	downloadURL, version, err := c.ensureDownloadableUrl()
	if err != nil {
		return fmt.Errorf("failed to ensure downloadable URL: %w", err)
	}

	archiveName := filepath.Base(downloadURL)

	tempFile, err := os.CreateTemp("", "tool-download-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	fmt.Printf("Downloading %s, version :: %s ...\n", toolName, version)
	if err := c.downloadFile(ctx, downloadURL, tempFile); err != nil {
		return fmt.Errorf("failed to download %s: %w", toolName, err)
	}

	if utils.IsArchive(downloadURL) {
		if err := utils.ExtractExecutable(tempFile.Name(), c.destPath, toolName, archiveName); err != nil {
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

	fmt.Printf("%s %s installed successfully at %s\n", toolName, version, c.destPath)
	return nil
}

// ensureDownloadableUrl ensures that the client has a direct downloadable URL.
// If the current URL is a GitHub API URL pointing to releases, it resolves it
// to a direct download URL using getReleaseDownloadURL.
//
// For "latest" version requests, it updates the version string with the actual
// resolved version number.
//
// Examples:
//   - URL that will be processed as GitHub API URL: "https://api.github.com/repos/owner/repo/releases/latest"
//   - URL that will be treated as direct download URL: "https://github.com/owner/repo/releases/download/v1.0/tool.tar.gz"
//
// Returns:
//   - The direct download URL
//   - The resolved version string
//   - Any error encountered during URL resolution
func (c *GithubClient) ensureDownloadableUrl() (string, string, error) {

	if !strings.Contains(c.DownloadUrl, "api.github.com") || !strings.Contains(c.DownloadUrl, "/releases/") {
		return c.DownloadUrl, c.version, nil
	}

	downloadURL, version, err := c.getReleaseDownloadURL(c.DownloadUrl, c.assetPattern)
	if err != nil {
		return "", "", fmt.Errorf("failed to get download URL: %w", err)
	}

	resolvedVersion := c.version
	if c.version == "latest" {
		resolvedVersion = version
	}

	return downloadURL, resolvedVersion, nil
}

// getReleaseDownloadURL fetches the download URL for a specific GitHub release
// getReleaseDownloadURL fetches release information from a GitHub repository API URL
// and returns the download URL for an asset that matches the provided regex pattern.
//
// Parameters:
//   - repoAPIURL: GitHub API URL for the repository release (e.g., "https://api.github.com/repos/owner/repo/releases/latest")
//   - assetPattern: Regular expression pattern to match against asset names
//
// Returns:
//   - The browser download URL of the matching asset
//   - The version of the release (with "v" prefix removed)
//   - An error if the request fails, JSON parsing fails, or no matching asset is found
//
// Errors:
//   - When the HTTP request to GitHub API fails
//   - When JSON decoding of response fails
//   - When no asset matching the provided pattern is found
func (c *GithubClient) getReleaseDownloadURL(repoAPIURL, assetPattern string) (string, string, error) {
	resp, err := c.httpClient.Get(repoAPIURL)
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch releases: %w", err)
	}
	defer resp.Body.Close()

	var release GithubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", "", fmt.Errorf("failed to parse release data: %w", err)
	}

	version := strings.TrimPrefix(release.TagName, "v")

	pattern := regexp.MustCompile(assetPattern)

	for _, asset := range release.Assets {
		if pattern.MatchString(asset.Name) {
			return asset.BrowserDownloadURL, version, nil
		}
	}

	return "", "", fmt.Errorf("could not find matching release asset with pattern: %s", assetPattern)
}

// downloadFile downloads content from the specified URL and writes it to the provided writer.
//
// Parameters:
//   - ctx: context.Context for controlling the request lifecycle
//   - url: the URL to download the file from
//   - w: io.Writer to which the downloaded content will be written
//
// Returns:
//   - error: if the request creation fails, the HTTP request fails,
//     the server returns a non-200 status code, or writing to the output fails
//
// The method uses the client's configured httpClient to perform the request.
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
