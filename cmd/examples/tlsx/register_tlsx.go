package main

import (
	"fmt"
	"log"

	"github.com/IxBahy/ASM/internal/scanners"
	"github.com/IxBahy/ASM/internal/scanners/tlsx"
)

func main() {
	registry := scanners.NewScannerRegistry()

	tlsxScanner := tlsx.NewTLSXScanner()

	// Register the scanner (setup is a no-op)
	registry.Register(tlsxScanner)

	// Display scanner info
	fmt.Printf("Scanner: %s\n", tlsxScanner.GetConfig().Name)
	fmt.Printf("Version: %s\n", tlsxScanner.GetInstallationState().Version)

	// Example scan
	target := "github.com"
	fmt.Printf("\nRunning TLS scan against %s...\n", target)

	result, err := tlsxScanner.Scan(target)
	if err != nil {
		log.Printf("Scan encountered errors: %v", err)
	}

	// Display results
	fmt.Println("\nScan Results:")
	for _, line := range result.Data {
		fmt.Println(line)
	}

	if len(result.Errors) > 0 {
		fmt.Println("\nErrors:")
		for _, err := range result.Errors {
			fmt.Println(" - " + err)
		}
	}

	fmt.Println("\nScan completed successfully")
}
