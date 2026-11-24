// Package backend provides core application services and state management for Vinta.
package backend

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Build-time configuration values injected via -ldflags for official releases.
var (
	buildFirebaseAPIKey     = ""
	buildFirebaseAuthDomain = ""
	buildFirebaseProjectID  = ""
	buildGoogleClientID     = ""
	buildGitHubClientID     = ""
)

// IntegrationCapabilities reports which optional OAuth and cloud backends are configured.
type IntegrationCapabilities struct {
	HasGoogleOAuth bool `json:"hasGoogleOAuth"`
	HasFirebase    bool `json:"hasFirebase"`
	HasGitHubOAuth bool `json:"hasGitHubOAuth"`
}

var (
	envLoadOnce sync.Once
	localEnvMap map[string]string
)

func loadEnvFileInto(path string, target map[string]string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.IndexByte(line, '=')
		if idx <= 0 {
			continue
		}
		k := strings.TrimSpace(line[:idx])
		v := strings.TrimSpace(line[idx+1:])
		if len(v) >= 2 && ((v[0] == '"' && v[len(v)-1] == '"') || (v[0] == '\'' && v[len(v)-1] == '\'')) {
			v = v[1 : len(v)-1]
		}
		if k != "" && v != "" {
			if _, exists := target[k]; !exists {
				target[k] = v
			}
		}
	}
}

func ensureLocalEnvLoaded() {
	envLoadOnce.Do(func() {
		localEnvMap = make(map[string]string)
		if cwd, err := os.Getwd(); err == nil {
			loadEnvFileInto(filepath.Join(cwd, ".env"), localEnvMap)
		}
		loadEnvFileInto(filepath.Join(GetAppDataDir(), ".env"), localEnvMap)
	})
}

// GetEnvConfig resolves a configuration key from process environment, local .env files, or build-time ldflags.
func GetEnvConfig(key string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	ensureLocalEnvLoaded()
	if v := strings.TrimSpace(localEnvMap[key]); v != "" {
		return v
	}
	switch key {
	case "VINTA_FIREBASE_API_KEY":
		return strings.TrimSpace(buildFirebaseAPIKey)
	case "VINTA_FIREBASE_AUTH_DOMAIN":
		return strings.TrimSpace(buildFirebaseAuthDomain)
	case "VINTA_FIREBASE_PROJECT_ID":
		return strings.TrimSpace(buildFirebaseProjectID)
	case "VINTA_GOOGLE_CLIENT_ID":
		return strings.TrimSpace(buildGoogleClientID)
	case "VINTA_GITHUB_CLIENT_ID":
		return strings.TrimSpace(buildGitHubClientID)
	default:
		return ""
	}
}

// GetIntegrationCapabilities returns which optional cloud and OAuth providers are available.
func (a *App) GetIntegrationCapabilities() IntegrationCapabilities {
	return IntegrationCapabilities{
		HasGoogleOAuth: GetEnvConfig("VINTA_GOOGLE_CLIENT_ID") != "",
		HasFirebase:    GetEnvConfig("VINTA_FIREBASE_API_KEY") != "" && GetEnvConfig("VINTA_FIREBASE_PROJECT_ID") != "",
		HasGitHubOAuth: GetEnvConfig("VINTA_GITHUB_CLIENT_ID") != "",
	}
}
