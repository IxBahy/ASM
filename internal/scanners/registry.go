package scanners

import (
	"fmt"
	"log"
	"sync"
)

type ScannerRegistry struct {
	scanners map[string]Scanner
	mu       sync.RWMutex
}

func NewScannerRegistry() *ScannerRegistry {
	return &ScannerRegistry{
		scanners: make(map[string]Scanner),
	}
}

func (r *ScannerRegistry) Register(scanner Scanner) {

	r.mu.Lock()
	defer r.mu.Unlock()

	config := scanner.GetConfig()
	if err := scanner.Setup(); err != nil {
		log.Fatalf("Failed to setup %s scanner: %v", config.Name, err)
	}

	r.scanners[config.Name] = scanner
	fmt.Printf("Scanner '%s' registered in registry\n", config.Name)

}

func (r *ScannerRegistry) Get(name string) (Scanner, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	scanner, exists := r.scanners[name]
	return scanner, exists
}

func (r *ScannerRegistry) GetAll() map[string]Scanner {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]Scanner, len(r.scanners))
	for k, v := range r.scanners {
		result[k] = v
	}

	return result
}
