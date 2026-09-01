// Package backend provides application icon extraction, caching, and resolution services.
package backend

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"Vinta/backend/services"
)

var (
	iconMemoryCache sync.Map
	iconFlightGroup sync.Map
)

type iconFlightCall struct {
	wg  sync.WaitGroup
	val string
	err error
}

func getIconCacheDir() string {
	dir := filepath.Join(GetAppDataDir(), "cache", "icons")
	_ = os.MkdirAll(dir, 0755)
	return dir
}

// computeIconHashKey generates a deterministic SHA-256 hash filename for a given app identifier.
func computeIconHashKey(sourceApp, aumid string) string {
	raw := strings.ToLower(strings.TrimSpace(sourceApp)) + "|" + strings.ToLower(strings.TrimSpace(aumid))
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// ResolveNativeAppIcon resolves the icon for sourceApp or aumid, caching in memory and persistent disk.
func (a *App) ResolveNativeAppIcon(sourceApp, aumid string) string {
	sApp := strings.TrimSpace(sourceApp)
	sAumid := strings.TrimSpace(aumid)

	if sApp == "" && sAumid == "" {
		return ""
	}

	hashKey := computeIconHashKey(sApp, sAumid)

	// 1. Check in-memory cache
	if val, ok := iconMemoryCache.Load(hashKey); ok {
		if str, ok := val.(string); ok {
			return str
		}
	}

	// 2. Request deduplication (singleflight) to prevent concurrent extractions of the same app
	callInterface, loaded := iconFlightGroup.LoadOrStore(hashKey, &iconFlightCall{})
	call := callInterface.(*iconFlightCall)

	if loaded {
		call.wg.Wait()
		return call.val
	}

	call.wg.Add(1)
	defer func() {
		iconFlightGroup.Delete(hashKey)
		call.wg.Done()
	}()

	cacheDir := getIconCacheDir()
	filePath := filepath.Join(cacheDir, hashKey+".png")

	// 3. Check persistent disk cache
	if info, err := os.Stat(filePath); err == nil && info.Size() > 0 {
		if data, err := os.ReadFile(filePath); err == nil {
			uri := "data:image/png;base64," + base64.StdEncoding.EncodeToString(data)
			iconMemoryCache.Store(hashKey, uri)
			call.val = uri
			return uri
		}
	}

	// 4. Native extraction via C++/WinRT bridge with 2-second timeout
	doneChan := make(chan bool, 1)
	go func() {
		extracted := services.ExtractAppIconToFile(sApp, sAumid, filePath, 128)
		doneChan <- extracted
	}()

	select {
	case success := <-doneChan:
		if success {
			if data, err := os.ReadFile(filePath); err == nil && len(data) > 0 {
				uri := "data:image/png;base64," + base64.StdEncoding.EncodeToString(data)
				iconMemoryCache.Store(hashKey, uri)
				call.val = uri
				return uri
			}
		}
	case <-time.After(2 * time.Second):
		// Extraction timed out; degrade gracefully
	}

	// Negative cache to prevent thrashing on unknown or invalid applications
	iconMemoryCache.Store(hashKey, "")
	call.val = ""
	return ""
}

// GetAppIcon is the Wails frontend binding to fetch the native application icon as a data URI.
func (a *App) GetAppIcon(sourceApp string, aumid string) string {
	return a.ResolveNativeAppIcon(sourceApp, aumid)
}
