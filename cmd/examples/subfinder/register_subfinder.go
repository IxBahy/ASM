package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/IxBahy/ASM/internal/scanners"
	"github.com/IxBahy/ASM/internal/scanners/subfinder"
)

func main() {
	registry := scanners.NewScannerRegistry()

	// Create the subfinder scanner
	subfinderScanner := subfinder.NewSubfinderScanner()

	// Register the scanner (setup is a no-op)
	registry.Register(subfinderScanner)

	// Display scanner info
	fmt.Printf("Scanner: %s\n", subfinderScanner.GetConfig().Name)
	fmt.Printf("Version: %s\n", subfinderScanner.GetInstallationState().Version)

	// Example scan - use a well-known domain
	target := "example.com"
	fmt.Printf("\nRunning Subfinder scan against %s...\n", target)
	fmt.Println("Please wait, subdomain enumeration may take a moment...")

	result, err := subfinderScanner.Scan(target)
	if err != nil {
		log.Printf("Scan encountered errors: %v", err)
	}

	// Display results
	if len(result.Data) > 0 {
		fmt.Println("\nScan Results:")

		// Parse the JSON output
		var subdomainResult subfinder.SubdomainResult
		if err := json.Unmarshal([]byte(result.Data[0]), &subdomainResult); err != nil {
			fmt.Printf("Error parsing results: %v\n", err)
			fmt.Println("Raw output:", result.Data[0])
		} else {
			fmt.Printf("Domain: %s\n", subdomainResult.Domain)
			fmt.Printf("Found %d subdomains:\n\n", len(subdomainResult.Subdomains))

			// Display subdomains in a formatted way
			for i, subdomain := range subdomainResult.Subdomains {
				fmt.Printf("%3d. %s\n", i+1, subdomain)
			}

			fmt.Printf("\nScan time: %s\n", subdomainResult.TimeStamp)
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
