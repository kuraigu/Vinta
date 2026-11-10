//go:build devtest

package backend

import "os"

// isTestModeSync returns true when compiled with -tags devtest unless suppressed via VINTA_NO_DEVTEST=1.
func isTestModeSync() bool {
	return os.Getenv("VINTA_NO_DEVTEST") != "1"
}
