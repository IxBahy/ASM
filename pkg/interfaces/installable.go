package interfaces

type Installable interface {
	Setup() error
	IsInstalled() bool
}
