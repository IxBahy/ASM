package main

import (
	"fmt"
	"log"

	"github.com/IxBahy/ASM/internal/scanners"
	"github.com/IxBahy/ASM/internal/scanners/wpscan"
)

func main() {
	registry := scanners.NewScannerRegistry()

	wpScanner := wpscan.NewWPScanScanner()

	fmt.Println("Setting up WPScan scanner...")
	if err := wpScanner.Setup(); err != nil {
		log.Fatalf("Failed to setup WPScan scanner: %v", err)
	}

	registry.Register(wpScanner)

	if !wpScanner.IsInstalled() {
		log.Fatal("WPScan installation could not be verified")
	}

	fmt.Printf("Scanner: %s\n", wpScanner.GetConfig().Name)
	fmt.Printf("Version: %s\n", wpScanner.GetInstallationState().Version)
	fmt.Printf("Path: %s\n", wpScanner.GetConfig().ExecutablePath)

	target := "https://wordpress.org"
	fmt.Printf("\nRunning WPScan against %s...\n", target)

	result, err := wpScanner.Scan(target)
	if err != nil {
		log.Fatalf("Scan encountered errors: %v", err)
	}

	if len(result.Errors) > 0 {
		fmt.Println("\nErrors encountered:")
		for _, err := range result.Errors {
			fmt.Println(" - " + err)
		}
	}

	if len(result.Data) > 0 {
		fmt.Println("\nResults:")

		resultLimit := 15
		for i, line := range result.Data {
			if i >= resultLimit {
				fmt.Printf(" - ... and %d more result lines\n", len(result.Data)-resultLimit)
				break
			}
			fmt.Println(" - " + line)
		}
	} else {
		fmt.Println("\nNo results found")
	}

	fmt.Println("\nScan completed successfully")
}
