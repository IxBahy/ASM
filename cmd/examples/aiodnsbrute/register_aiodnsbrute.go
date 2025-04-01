package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/IxBahy/ASM/internal/scanners"
	"github.com/IxBahy/ASM/internal/scanners/aiodnsbrute"
)

func main() {
	registry := scanners.NewScannerRegistry()

	// Create the aiodnsbrute scanner
	dnsScanner := aiodnsbrute.NewAioDNSBruteScanner()

	fmt.Println("Setting up AioDNSBrute scanner...")
	if err := dnsScanner.Setup(); err != nil {
		log.Fatalf("Failed to setup AioDNSBrute scanner: %v", err)
	}

	registry.Register(dnsScanner)

	if !dnsScanner.IsInstalled() {
		log.Fatal("AioDNSBrute installation could not be verified")
	}

	fmt.Printf("Scanner: %s\n", dnsScanner.GetConfig().Name)
	fmt.Printf("Version: %s\n", dnsScanner.GetInstallationState().Version)
	fmt.Printf("Path: %s\n", dnsScanner.GetConfig().ExecutablePath)

	// Example scan
	target := "buguard.io"
	fmt.Printf("\nRunning AioDNSBrute scan against %s...\n", target)
	fmt.Println("Please wait, DNS brute-forcing may take a moment...")

	result, err := dnsScanner.Scan(target)
	if err != nil {
		fmt.Printf("Scan encountered errors: %v\n", err)
	}

	// Display results
	if len(result.Data) > 0 {
		fmt.Println("\nScan Results:")

		// Parse the JSON output
		var bruteResult aiodnsbrute.DNSBruteResult
		if err := json.Unmarshal([]byte(result.Data[0]), &bruteResult); err != nil {
			fmt.Printf("Error parsing results: %v\n", err)
			fmt.Println("Raw output:", result.Data[0])
		} else {
			fmt.Printf("Target domain: %s\n", bruteResult.Domain)

			if len(bruteResult.Subdomains) > 0 {
				fmt.Printf("Found %d subdomains:\n\n", len(bruteResult.Subdomains))

				for i, subdomain := range bruteResult.Subdomains {
					fmt.Printf(" %3d. %s\n", i+1, subdomain)
				}
			} else {
				fmt.Println("No subdomains found")
			}

			fmt.Printf("\nScan time: %s\n", bruteResult.TimeStamp)
		}
	} else {
		fmt.Println("\nNo subdomains found.")
	}

	if len(result.Errors) > 0 {
		fmt.Println("\nErrors:")
		for _, err := range result.Errors {
			fmt.Println(" - " + err)
		}
	}

	fmt.Println("\nScan completed successfully")
}
