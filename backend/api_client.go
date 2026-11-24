// Package backend provides core application services and state management for Vinta.
package backend

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type sfCall struct {
	wg  sync.WaitGroup
	val []byte
	err error
}

type singleflightGroup struct {
	mu    sync.Mutex
	calls map[string]*sfCall
}

func (g *singleflightGroup) Do(key string, fn func() ([]byte, error)) ([]byte, error) {
	g.mu.Lock()
	if g.calls == nil {
		g.calls = make(map[string]*sfCall)
	}
	if c, ok := g.calls[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}
	c := &sfCall{}
	c.wg.Add(1)
	g.calls[key] = c
	g.mu.Unlock()

	c.val, c.err = fn()
	c.wg.Done()

	g.mu.Lock()
	delete(g.calls, key)
	g.mu.Unlock()

	return c.val, c.err
}

type apiCacheEntry struct {
	ExpiresAt int64  `json:"expiresAt"`
	Body      []byte `json:"body"`
}

var (
	sharedAPIClient = &http.Client{
		Timeout: 15 * time.Second,
	}
	sharedAPISingleflight singleflightGroup
	apiBodyBufPool        = sync.Pool{
		New: func() any {
			return bytes.NewBuffer(make([]byte, 0, 16*1024))
		},
	}
)

func getAPICachePath(cacheKey string) string {
	sum := sha256.Sum256([]byte(cacheKey))
	return filepath.Join(GetAppDataDir(), "cache", "api", hex.EncodeToString(sum[:])+".json")
}

func readAPIDiskCache(cacheKey string) ([]byte, bool) {
	raw, err := os.ReadFile(getAPICachePath(cacheKey))
	if err != nil {
		return nil, false
	}
	var entry apiCacheEntry
	if err := json.Unmarshal(raw, &entry); err != nil {
		return nil, false
	}
	if time.Now().Unix() >= entry.ExpiresAt {
		return nil, false
	}
	return entry.Body, true
}

func writeAPIDiskCache(cacheKey string, ttl time.Duration, body []byte) {
	if ttl <= 0 {
		return
	}
	p := getAPICachePath(cacheKey)
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return
	}
	entry := apiCacheEntry{
		ExpiresAt: time.Now().Add(ttl).Unix(),
		Body:      body,
	}
	if encoded, err := json.Marshal(entry); err == nil {
		_ = os.WriteFile(p, encoded, 0600)
	}
}

func cachedHTTPDo(
	ctx context.Context,
	cacheKey string,
	ttl time.Duration,
	forceRefresh bool,
	buildReq func(context.Context) (*http.Request, error),
) ([]byte, error) {
	if !forceRefresh && ttl > 0 {
		if body, ok := readAPIDiskCache(cacheKey); ok {
			return body, nil
		}
	}

	return sharedAPISingleflight.Do(cacheKey, func() ([]byte, error) {
		if !forceRefresh && ttl > 0 {
			if body, ok := readAPIDiskCache(cacheKey); ok {
				return body, nil
			}
		}
		req, err := buildReq(ctx)
		if err != nil {
			return nil, err
		}
		resp, err := sharedAPIClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer func() { _ = resp.Body.Close() }()

		buf := apiBodyBufPool.Get().(*bytes.Buffer)
		buf.Reset()
		_, err = io.Copy(buf, io.LimitReader(resp.Body, 4*1024*1024))
		body := append([]byte(nil), buf.Bytes()...)
		if buf.Cap() <= 512*1024 {
			buf.Reset()
			apiBodyBufPool.Put(buf)
		}
		if err != nil {
			return nil, err
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, fmt.Errorf("upstream status %d: %s", resp.StatusCode, string(body))
		}
		writeAPIDiskCache(cacheKey, ttl, body)
		return body, nil
	})
}
