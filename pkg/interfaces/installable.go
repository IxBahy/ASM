package interfaces

type Installable interface {
	Setup() error
	IsInstalled() bool
	GetInstallationState() InstallationState
}
type InstallationState struct {
	Installed bool
	Version   string
	Path      string
	Errors    []string
}
