//go:build devtest

package backend

// isTestModeSync returns true when compiled with -tags devtest.
func isTestModeSync() bool {
	return true
}
