package main

import (
	"fmt"
	"log"

	"github.com/IxBahy/ASM/internal/scanners"
	"github.com/IxBahy/ASM/internal/scanners/nmap"
)

func main() {
	registry := scanners.NewScannerRegistry()
	nmapScanner := nmap.NewNmapScanner()

	registry.Register(nmapScanner)

	if nmapScanner.IsInstalled() {
		result, err := nmapScanner.Scan("localhost")
		if err != nil {
			log.Printf("Scan error: %v", err)
		}

		fmt.Println("Scan results:")
		for _, line := range result.Data {
			fmt.Println(line)
		}
	}
}
