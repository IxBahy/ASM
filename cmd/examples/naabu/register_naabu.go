package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/IxBahy/ASM/internal/scanners"
	"github.com/IxBahy/ASM/internal/scanners/naabu"
)

func main() {
	registry := scanners.NewScannerRegistry()

	// Create the naabu scanner
	naabuScanner := naabu.NewNaabuScanner()

	// Register the scanner (setup is a no-op for embedded scanners)
	registry.Register(naabuScanner)

	// Display scanner info
	fmt.Printf("Scanner: %s\n", naabuScanner.GetConfig().Name)
	fmt.Printf("Version: %s\n", naabuScanner.GetInstallationState().Version)

	// Example scan - use a well-known domain
	target := "example.com"
	fmt.Printf("\nRunning Naabu port scan against %s...\n", target)
	fmt.Println("Please wait, port scanning may take a moment...")

	result, err := naabuScanner.Scan(target)
	if err != nil {
		log.Printf("Scan encountered errors: %v", err)
	}

	// Display results
	if len(result.Data) > 0 {
		fmt.Println("\nScan Results:")

		// Parse the JSON output
		var scanResult naabu.PortScanResult
		if err := json.Unmarshal([]byte(result.Data[0]), &scanResult); err != nil {
			fmt.Printf("Error parsing results: %v\n", err)
			fmt.Println("Raw output:", result.Data[0])
		} else {
			fmt.Printf("Target: %s\n", scanResult.Host)

			if len(scanResult.Ports) > 0 {
				fmt.Printf("Found %d open ports:\n\n", len(scanResult.Ports))

				for i, port := range scanResult.Ports {
					fmt.Printf(" %3d. Port %s/tcp\n", i+1, port)
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
