package scanners

type Scanner interface {
	Name() string
	Install() error
	Scan(target string) (Result, error)
}

type Result struct {
	Vulnerabilities []string
	Errors          []string
}
