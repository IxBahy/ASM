package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/IxBahy/ASM/internal/scanners"
	"github.com/IxBahy/ASM/internal/scanners/trufflehog"
)

// Finding structure for the Go version of TruffleHog
type Finding struct {
	DetectorName string `json:"DetectorName"`
	DetectorType int    `json:"DetectorType"`
	Raw          string `json:"Raw"`
	Redacted     string `json:"Redacted"`
	File         string `json:"SourceMetadata"`
	SourceID     string `json:"SourceID"`
}

func main() {
	registry := scanners.NewScannerRegistry()

	truffleScanner := trufflehog.NewTruffleHogScanner()

	fmt.Println("Setting up TruffleHog scanner...")
	if err := truffleScanner.Setup(); err != nil {
		log.Fatalf("Failed to setup TruffleHog scanner: %v", err)
	}

	registry.Register(truffleScanner)

	if !truffleScanner.IsInstalled() {
		log.Fatal("TruffleHog installation could not be verified")
	}

	fmt.Printf("Scanner: %s\n", truffleScanner.GetConfig().Name)
	fmt.Printf("Version: %s\n", truffleScanner.GetInstallationState().Version)
	fmt.Printf("Path: %s\n", truffleScanner.GetConfig().ExecutablePath)

	// Create a demo repository with test secrets
	tempDir, err := createDemoRepo()
	if err != nil {
		log.Fatalf("Failed to create demo repo: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fmt.Printf("\nRunning TruffleHog scan against demo repo at %s...\n", tempDir)

	result, err := truffleScanner.Scan(tempDir)
	if err != nil {
		fmt.Printf("Scan encountered errors: %v\n", err)
	}

	// Display findings
	fmt.Println("\nScan Results:")

	if len(result.Data) > 0 {
		fmt.Printf("Found %d potential secrets:\n\n", len(result.Data))

		for i, finding := range result.Data {
			var decodedFinding Finding
			if err := json.Unmarshal([]byte(finding), &decodedFinding); err != nil {
				fmt.Printf(" - Raw result: %s\n", finding)
				continue
			}

			// Truncate very long secrets
			rawSecret := decodedFinding.Raw
			secretDisplay := rawSecret
			if len(secretDisplay) > 40 {
				secretDisplay = secretDisplay[:20] + "..." + secretDisplay[len(secretDisplay)-17:]
			}

			fmt.Printf(" %d. [%s]\n", i+1, decodedFinding.DetectorName)
			fmt.Printf("    Source: %s\n", decodedFinding.SourceID)
			fmt.Printf("    Secret: %s\n\n", secretDisplay)
		}
	} else {
		fmt.Println("No secrets found")
	}

	fmt.Println("\nScan completed successfully")
}

// Create a temporary repository with test secrets for demonstration
func createDemoRepo() (string, error) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "trufflehog-demo-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	// Create a test file with a fake secret
	secretFile := filepath.Join(tempDir, "config.js")
	secretContent := `
// Configuration file
const config = {
  apiKey: "1234567890abcdef1234567890abcdef",
  dbPassword: "SuperSecretP@ssword123!",
  awsSecret: "AKIAIOSFODNN7EXAMPLE",
  token: "ghp_aBcDeFgHiJkLmNoPqRsTuVwXyZ0123456789",
  privateKey: "-----BEGIN RSA PRIVATE KEY-----\nkdfjlsdjflksdjflkdsjflksdjf\n-----END RSA PRIVATE KEY-----"
};

module.exports = config;
`

	if err := os.WriteFile(secretFile, []byte(secretContent), 0644); err != nil {
		os.RemoveAll(tempDir)
		return "", fmt.Errorf("failed to create test file: %w", err)
	}

	return tempDir, nil
}
