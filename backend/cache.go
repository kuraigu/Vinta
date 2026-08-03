package backend

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// Cache provides a local file caching mechanism for remote resources.
type Cache struct {
	cacheDir string
}

// NewCache initializes a new Cache instance and ensures the cache directory exists.
func NewCache() *Cache {
	localAppData := os.Getenv("LOCALAPPDATA")
	var cacheDir string
	if localAppData != "" {
		cacheDir = filepath.Join(localAppData, "Vinta", "cache")
	} else {
		cacheDir = filepath.Join(os.TempDir(), "Vinta", "cache")
	}

	_ = os.MkdirAll(cacheDir, 0755)

	return &Cache{
		cacheDir: cacheDir,
	}
}

func downloadFile(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	_, err = io.Copy(out, resp.Body)
	return err
}

// CacheIcons takes a list of URLs, downloads missing ones, and returns a map of URL to a local custom URI.
func (c *Cache) CacheIcons(urls []string) map[string]string {
	results := make(map[string]string)

	for _, u := range urls {
		u = strings.TrimSpace(u)
		if u == "" || !strings.HasPrefix(u, "http") {
			continue
		}

		// Generate a safe unique filename based on the URL hash
		hash := sha256.Sum256([]byte(u))
		hashStr := hex.EncodeToString(hash[:])

		ext := filepath.Ext(u)
		if ext == "" {
			ext = ".svg"
		} else {
			if idx := strings.Index(ext, "?"); idx != -1 {
				ext = ext[:idx]
			}
		}

		filename := hashStr + ext
		localPath := filepath.Join(c.cacheDir, filename)

		// Check if file exists, download if it doesn't
		if _, err := os.Stat(localPath); os.IsNotExist(err) {
			if err := downloadFile(u, localPath); err != nil {
				continue
			}
		}

		// Return a custom URL path for the Wails AssetsHandler to intercept
		results[u] = "/icon-cache/" + filename
	}

	return results
}

// ServeHTTP implements the http.Handler interface to serve cached files directly to the frontend.
func (c *Cache) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if strings.HasPrefix(req.URL.Path, "/icon-cache/") {
		// Strip the prefix to get the actual filename
		filename := strings.TrimPrefix(req.URL.Path, "/icon-cache/")
		fullPath := filepath.Join(c.cacheDir, filename)

		// http.ServeFile automatically handles reading, MIME types, and 404s if the file is missing
		http.ServeFile(res, req, fullPath)
		return
	}

	if strings.HasPrefix(req.URL.Path, "/local/") {
		localPath := strings.TrimPrefix(req.URL.Path, "/local/")
		println("DEBUG CACHE: Original localPath:", localPath)
		if unescaped, err := url.PathUnescape(localPath); err == nil {
			localPath = unescaped
		}
		println("DEBUG CACHE: Unescaped localPath:", localPath)
		if localPath != "" {
			http.ServeFile(res, req, localPath)
			return
		}
	}

	// If it is not an icon request, return 404 so Wails knows to handle it normally
	res.WriteHeader(http.StatusNotFound)
}
