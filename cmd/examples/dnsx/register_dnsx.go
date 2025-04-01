package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/IxBahy/ASM/internal/scanners"
	"github.com/IxBahy/ASM/internal/scanners/dnsx"
)

func main() {
	registry := scanners.NewScannerRegistry()

	// Create the DNSx scanner
	dnsxScanner := dnsx.NewDNSxScanner()

	// Register the scanner
	registry.Register(dnsxScanner)

	// Display scanner info
	fmt.Printf("Scanner: %s\n", dnsxScanner.GetConfig().Name)
	fmt.Printf("Version: %s\n", dnsxScanner.GetInstallationState().Version)

	// Example scan
	target := "example.com"
	fmt.Printf("\nRunning DNS lookup for %s...\n", target)

	result, err := dnsxScanner.Scan(target)
	if err != nil {
		log.Fatalf("Scan failed: %v", err)
	}

	// Display results
	if len(result.Data) > 0 {
		fmt.Println("\nScan Results:")

		// Parse the JSON output
		var dnsResult dnsx.DNSResult
		if err := json.Unmarshal([]byte(result.Data[0]), &dnsResult); err != nil {
			fmt.Printf("Error parsing results: %v\n", err)
			fmt.Println("Raw output:", result.Data[0])
		} else {
			fmt.Printf("Target domain: %s\n", dnsResult.Domain)

			// Display IP addresses
			if len(dnsResult.IPs) > 0 {
				fmt.Printf("\nFound %d IP addresses:\n", len(dnsResult.IPs))
				for i, ip := range dnsResult.IPs {
					fmt.Printf("  %d. %s\n", i+1, ip)
				}
			} else {
				fmt.Println("\nNo IP addresses found")
			}

			fmt.Printf("\nScan time: %s\n", dnsResult.TimeStamp)
		}
	} else {
		fmt.Println("\nNo DNS records found.")
	}

	// Display any errors
	if len(result.Errors) > 0 {
		fmt.Println("\nErrors:")
		for _, err := range result.Errors {
			fmt.Println(" - " + err)
		}
	}

	fmt.Println("\nScan completed successfully")
}
