package scanners

type ScannerConfig struct {
	Name           string
	Version        string
	InstallLink    string
	ExecutablePath string
	Base_Command   string
}

type Result struct {
	Data   []string
	Errors []string
}
