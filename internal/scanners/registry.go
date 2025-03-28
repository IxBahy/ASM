package scanners

import (
	"fmt"
	"sync"
)

// ScannerRegistry maintains a registry of available scanners
type ScannerRegistry struct {
	scanners map[string]Scanner
	mu       sync.RWMutex
}

// NewScannerRegistry creates a new scanner registry
func NewScannerRegistry() *ScannerRegistry {
	return &ScannerRegistry{
		scanners: make(map[string]Scanner),
	}
}

// Register adds a scanner to the registry
func (r *ScannerRegistry) Register(scanner Scanner) {
	r.mu.Lock()
	defer r.mu.Unlock()

	config := scanner.Config()
	r.scanners[config.Name] = scanner
	fmt.Printf("Scanner '%s' registered\n", config.Name)
}

// Get retrieves a scanner by name
func (r *ScannerRegistry) Get(name string) (Scanner, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	scanner, ok := r.scanners[name]
	if !ok {
		return nil, fmt.Errorf("scanner '%s' not found", name)
	}

	return scanner, nil
}

// List returns all registered scanners
func (r *ScannerRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var names []string
	for name := range r.scanners {
		names = append(names, name)
	}

	return names
}
