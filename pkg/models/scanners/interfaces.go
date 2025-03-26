package scanners

type Scanner interface {
	Config() ScannerConfig
	Setup() error
	Scan(target string) (Result, error)
}
