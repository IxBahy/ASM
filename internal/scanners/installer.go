package scanners

import (
	"errors"
)

func InstallScanner(s Scanner) error {
	if s == nil {
		return errors.New("no scanner was provided")
	}
	return s.Install()
}
