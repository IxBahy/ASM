package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/IxBahy/ASM/internal/scanners"
	"github.com/IxBahy/ASM/internal/scanners/katana"
)

func main() {
	registry := scanners.NewScannerRegistry()

	// Create the katana scanner
	katanaScanner := katana.NewKatanaScanner()

	// Register the scanner (no setup needed for embedded scanners)
	registry.Register(katanaScanner)

	// Display scanner info
	fmt.Printf("Scanner: %s\n", katanaScanner.GetConfig().Name)
	fmt.Printf("Version: %s\n", katanaScanner.GetInstallationState().Version)

	// Example scan - use a well-known domain
	target := "https://example.com"
	fmt.Printf("\nRunning Katana web crawler against %s...\n", target)
	fmt.Println("Please wait, crawling may take a moment...")

	result, err := katanaScanner.Scan(target)
	if err != nil {
		log.Printf("Scan encountered errors: %v", err)
	}

	// Display results
	if len(result.Data) > 0 {
		fmt.Println("\nScan Results:")

		// Parse the JSON output
		var crawlResult katana.CrawlResult
		if err := json.Unmarshal([]byte(result.Data[0]), &crawlResult); err != nil {
			fmt.Printf("Error parsing results: %v\n", err)
			fmt.Println("Raw output:", result.Data[0])
		} else {
			fmt.Printf("Target domain: %s\n", crawlResult.Domain)
			fmt.Printf("Found %d URLs in %s\n\n", crawlResult.Count, crawlResult.ElapsedTime)

			// Display URLs and their details
			maxDisplay := 20
			if crawlResult.Count < maxDisplay {
				maxDisplay = crawlResult.Count
			}

			for i := 0; i < maxDisplay; i++ {
				if i >= len(crawlResult.URLs) {
					break
				}

				url := crawlResult.URLs[i]
				fmt.Printf(" %3d. %s\n", i+1, url.URL)
				fmt.Printf("      Status: %d | Type: %s | Depth: %d\n",
					url.StatusCode, url.ContentType, url.Depth)

				if len(url.Parameters) > 0 {
					fmt.Printf("      Parameters: ")
					paramCount := 0
					for k, v := range url.Parameters {
						if paramCount > 0 {
							fmt.Print(", ")
						}
						fmt.Printf("%s=%s", k, v)
						paramCount++
					}
					fmt.Println()
				}
				fmt.Println()
			}

			if crawlResult.Count > maxDisplay {
				fmt.Printf("... and %d more URLs\n", crawlResult.Count-maxDisplay)
			}

			fmt.Printf("\nScan time: %s\n", crawlResult.TimeStamp)
		}
	} else {
		fmt.Println("\nNo URLs found.")
	}

	if len(result.Errors) > 0 {
		fmt.Println("\nErrors:")
		for _, err := range result.Errors {
			fmt.Println(" - " + err)
		}
	}

	fmt.Println("\nScan completed successfully")
}
