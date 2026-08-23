//go:build !devtest

package backend

// isTestModeSync returns false for release and standard builds (compiled out at compile time).
func isTestModeSync() bool {
	return false
}
