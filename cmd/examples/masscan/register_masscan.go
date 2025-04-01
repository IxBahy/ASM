package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/IxBahy/ASM/internal/scanners"
	"github.com/IxBahy/ASM/internal/scanners/masscan"
)

func main() {
	registry := scanners.NewScannerRegistry()

	masscanScanner := masscan.NewMassScanScanner()

	fmt.Println("Setting up MassScan scanner...")
	if err := masscanScanner.Setup(); err != nil {
		log.Fatalf("Failed to setup MassScan scanner: %v", err)
	}

	registry.Register(masscanScanner)

	if !masscanScanner.IsInstalled() {
		log.Fatal("MassScan installation could not be verified")
	}

	fmt.Printf("Scanner: %s\n", masscanScanner.GetConfig().Name)
	fmt.Printf("Version: %s\n", masscanScanner.GetInstallationState().Version)
	fmt.Printf("Path: %s\n", masscanScanner.GetConfig().ExecutablePath)

	// Example scan
	target := "example.com"
	fmt.Printf("\nRunning MassScan port scan against %s...\n", target)
	fmt.Println("Please wait, port scanning may take a moment...")

	result, err := masscanScanner.Scan(target)
	if err != nil {
		fmt.Printf("Scan encountered errors: %v\n", err)
	}

	// Display results
	if len(result.Data) > 0 {
		fmt.Println("\nScan Results:")

		// Parse the JSON output
		var scanResult masscan.MassScanResult
		if err := json.Unmarshal([]byte(result.Data[0]), &scanResult); err != nil {
			fmt.Printf("Error parsing results: %v\n", err)
			fmt.Println("Raw output:", result.Data[0])
		} else {
			fmt.Printf("Target: %s (%s)\n", scanResult.Host, scanResult.IP)

			if len(scanResult.Ports) > 0 {
				fmt.Printf("Found %d open ports:\n\n", len(scanResult.Ports))

				for i, port := range scanResult.Ports {
					fmt.Printf(" %3d. Port %s\n", i+1, port)
				}
			} else {
				fmt.Println("No open ports found")
			}

			fmt.Printf("\nScan time: %s\n", scanResult.TimeStamp)
		}
	} else {
		fmt.Println("\nNo ports found.")
	}

	if len(result.Errors) > 0 {
		fmt.Println("\nErrors:")
		for _, err := range result.Errors {
			fmt.Println(" - " + err)
		}
	}

	fmt.Println("\nScan completed successfully")
}
