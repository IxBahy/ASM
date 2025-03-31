package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/IxBahy/ASM/internal/scanners"
	"github.com/IxBahy/ASM/internal/scanners/sqlmap"
)

func main() {
	registry := scanners.NewScannerRegistry()

	sqlmapScanner := sqlmap.NewSQLMapScanner()

	fmt.Println("Setting up SQLMap scanner using Python pip...")
	if err := sqlmapScanner.Setup(); err != nil {
		log.Fatalf("Failed to setup SQLMap scanner: %v", err)
	}

	registry.Register(sqlmapScanner)

	if !sqlmapScanner.IsInstalled() {
		log.Fatal("SQLMap installation could not be verified")
	}

	fmt.Printf("Scanner: %s\n", sqlmapScanner.GetConfig().Name)
	fmt.Printf("Version: %s\n", sqlmapScanner.GetInstallationState().Version)
	fmt.Printf("Path: %s\n", sqlmapScanner.GetConfig().ExecutablePath)

	// Example scan against a test site
	target := "http://testphp.vulnweb.com/listproducts.php?cat=1"
	fmt.Printf("\nRunning SQLMap scan against %s...\n", target)

	result, err := sqlmapScanner.Scan(target)
	if err != nil {
		fmt.Printf("Scan encountered errors: %v\n", err)
	}

	// Display summarized results
	fmt.Println("\nScan Results Summary:")

	vulnFound := false
	for _, line := range result.Data {
		if strings.Contains(line, "is vulnerable") ||
			strings.Contains(line, "Parameter") && strings.Contains(line, "injectable") {
			fmt.Printf(" [!] %s\n", line)
			vulnFound = true
		}
	}

	if !vulnFound {
		fmt.Println(" [+] No SQL injection vulnerabilities were detected")
	}

	fmt.Println("\nScan completed successfully")
}
