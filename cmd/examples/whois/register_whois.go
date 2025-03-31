package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/IxBahy/ASM/internal/scanners"
	"github.com/IxBahy/ASM/internal/scanners/whois"
)

func main() {
	registry := scanners.NewScannerRegistry()

	whoisScanner := whois.NewWhoisScanner()

	// Setup the scanner (install if not already installed)
	fmt.Println("Setting up WHOIS scanner...")
	if err := whoisScanner.Setup(); err != nil {
		log.Fatalf("Failed to setup WHOIS scanner: %v", err)
	}

	registry.Register(whoisScanner)

	// Check if installed successfully
	if !whoisScanner.IsInstalled() {
		log.Fatal("WHOIS installation could not be verified")
	}

	// Display scanner info
	fmt.Printf("Scanner: %s\n", whoisScanner.GetConfig().Name)
	fmt.Printf("Version: %s\n", whoisScanner.GetInstallationState().Version)
	fmt.Printf("Path: %s\n", whoisScanner.GetConfig().ExecutablePath)

	// Example scan
	target := "example.com"
	fmt.Printf("\nRunning WHOIS lookup for %s...\n", target)

	result, err := whoisScanner.Scan(target)
	if err != nil {
		log.Printf("Scan encountered errors: %v", err)
	}

	// Extract key information from WHOIS output
	var registrar, creationDate, expiryDate string

	for _, line := range result.Data {
		lowerLine := strings.ToLower(line)
		if strings.Contains(lowerLine, "registrar:") {
			registrar = strings.TrimSpace(strings.Split(line, ":")[1])
		} else if strings.Contains(lowerLine, "creation date:") {
			creationDate = strings.TrimSpace(strings.Split(line, ":")[1])
		} else if strings.Contains(lowerLine, "registry expiry date:") {
			expiryDate = strings.TrimSpace(strings.Split(line, ":")[1])
		}
	}

	fmt.Println("\nDomain Information:")
	fmt.Printf("  Registrar: %s\n", registrar)
	fmt.Printf("  Creation Date: %s\n", creationDate)
	fmt.Printf("  Expiry Date: %s\n", expiryDate)

	// Display full output if requested
	fmt.Println("\nFull WHOIS data:")
	maxLines := 20
	for i, line := range result.Data {
		if i >= maxLines {
			fmt.Printf("  ... and %d more lines\n", len(result.Data)-maxLines)
			break
		}
		fmt.Printf("  %s\n", line)
	}

	fmt.Println("\nScan completed successfully")
}
