package scanners

import (
	"fmt"
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

func (r *ScannerRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var names []string
	for name := range r.scanners {
		names = append(names, name)
	}

	return names
}
