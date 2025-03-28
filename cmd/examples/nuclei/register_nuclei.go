package main

import (
	"fmt"
	"log"

	"github.com/IxBahy/ASM/internal/scanners"
	"github.com/IxBahy/ASM/internal/scanners/nuclei"
)

func main() {
	registry := scanners.NewScannerRegistry()

	nucleiScanner := nuclei.NewNucleiScanner()

	if err := nucleiScanner.Setup(); err != nil {
		log.Fatalf("Failed to setup Nuclei scanner: %v", err)
	}

	registry.Register(nucleiScanner)

	if nucleiScanner.IsInstalled() {
		fmt.Println("Running Nuclei scan on example target...")

		result, err := nucleiScanner.Scan("example.com")
		if err != nil {
			log.Printf("Scan error: %v", err)
			if len(result.Errors) > 0 {
				fmt.Println("Error details:")
				for _, errLine := range result.Errors {
					fmt.Println(errLine)
				}
			}
		} else {
			fmt.Println("Scan results:")
			for _, line := range result.Data {
				fmt.Println(line)
			}
		}
	}
}
