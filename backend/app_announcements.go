// Package backend provides core application services and offline caching for Vinta.
package backend

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// AnnouncementItem represents a developer announcement or major release highlight.
type AnnouncementItem struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Body         string `json:"body"`
	URL          string `json:"url"`
	PublishedAt  string `json:"publishedAt"`
	AuthorName   string `json:"authorName"`
	AuthorAvatar string `json:"authorAvatar"`
	IsUnread     bool   `json:"isUnread"`
}

// AnnouncementsPayload encapsulates fetched announcements and unread state.
type AnnouncementsPayload struct {
	Items     []AnnouncementItem `json:"items"`
	HasUnread bool               `json:"hasUnread"`
}

// resolveCacheDir returns and creates %APPDATA%/FreeMatrix Studios/Vinta/cache.
func resolveCacheDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to locate user config directory: %w", err)
	}
	cacheDir := filepath.Join(configDir, "FreeMatrix Studios", "Vinta", "cache")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}
	return cacheDir, nil
}

// GetReleaseNotes retrieves release changelog markdown, prioritizing local disk cache.
func (a *App) GetReleaseNotes(version string) (string, error) {
	cleanVersion := strings.TrimSpace(version)
	if cleanVersion == "" {
		cleanVersion = AppVersion
	}
	cleanVersion = strings.TrimPrefix(cleanVersion, "v")
	if cleanVersion == "" {
		cleanVersion = "0.1.0"
	}

	cacheDir, err := resolveCacheDir()
	if err == nil {
		cacheFile := filepath.Join(cacheDir, fmt.Sprintf("release_%s.json", cleanVersion))
		if data, err := os.ReadFile(cacheFile); err == nil && len(data) > 0 {
			var cached struct {
				Body string `json:"body"`
			}
			if err := json.Unmarshal(data, &cached); err == nil && cached.Body != "" {
				return cached.Body, nil
			}
		}
	}

	tag := cleanVersion
	if !strings.HasPrefix(tag, "v") {
		tag = "v" + tag
	}

	apiURL := fmt.Sprintf("https://api.github.com/repos/kuraigu/Vinta/releases/tags/%s", tag)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return a.bundledReleaseNotes(cleanVersion), nil
	}
	req.Header.Set("User-Agent", "Vinta-App")
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if token := resolveUpdaterToken(); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return a.bundledReleaseNotes(cleanVersion), nil
	}
	defer func() { _ = resp.Body.Close() }()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return a.bundledReleaseNotes(cleanVersion), nil
	}

	var ghRelease struct {
		Body string `json:"body"`
	}
	if err := json.Unmarshal(bodyBytes, &ghRelease); err != nil || strings.TrimSpace(ghRelease.Body) == "" {
		return a.bundledReleaseNotes(cleanVersion), nil
	}

	if cacheDir != "" && ghRelease.Body != "" {
		cacheFile := filepath.Join(cacheDir, fmt.Sprintf("release_%s.json", cleanVersion))
		payload, _ := json.Marshal(map[string]string{"body": ghRelease.Body})
		_ = os.WriteFile(cacheFile, payload, 0600)
	}

	return ghRelease.Body, nil
}

func (a *App) bundledReleaseNotes(version string) string {
	return fmt.Sprintf(`## Vinta v%s

Welcome to Vinta! The modern interactive dynamic island utility designed for Windows.

### Built-in Highlights
- **Interactive Island**: Top-docked expanding widget surface with smooth window animation.
- **Home Glance Suite**: Live Weather, Calendar glance, Action Center quick settings, and System Telemetry.
- **Developer Announcements**: Live updates and release notes synchronized directly from GitHub.
- **Audio & Media Island**: Real-time volume mixer, media visualizers, and playback controls.

*Official online release notes and changelogs will automatically synchronize when connected to GitHub.*`, version)
}

// Atom XML structures for GitHub Discussions Announcements feed.
type atomFeed struct {
	XMLName xml.Name    `xml:"feed"`
	Entries []atomEntry `xml:"entry"`
}

type atomEntry struct {
	ID        string     `xml:"id"`
	Title     string     `xml:"title"`
	Published string     `xml:"published"`
	Link      atomLink   `xml:"link"`
	Content   string     `xml:"content"`
	Author    atomAuthor `xml:"author"`
}

type atomLink struct {
	Href string `xml:"href,attr"`
}

type atomAuthor struct {
	Name string `xml:"name"`
	URI  string `xml:"uri"`
}

// GetDeveloperAnnouncements fetches official developer announcements with hash-based unread detection.
func (a *App) GetDeveloperAnnouncements() (AnnouncementsPayload, error) {
	cacheDir, _ := resolveCacheDir()
	var hashFile, cacheFile string
	if cacheDir != "" {
		hashFile = filepath.Join(cacheDir, "announcement_hash.txt")
		cacheFile = filepath.Join(cacheDir, "announcements.json")
	}

	client := &http.Client{Timeout: 5 * time.Second}

	// 1. Primary Source: GitHub Discussions Announcements Atom Feed
	discussionsURL := "https://github.com/kuraigu/Vinta/discussions/categories/announcements.atom"
	reqDisc, err := http.NewRequest("GET", discussionsURL, nil)
	if err == nil {
		reqDisc.Header.Set("User-Agent", "Vinta-App")
		if token := resolveUpdaterToken(); token != "" {
			reqDisc.Header.Set("Authorization", "Bearer "+token)
		}
		if respDisc, err := client.Do(reqDisc); err == nil {
			defer func() { _ = respDisc.Body.Close() }()
			if respDisc.StatusCode == http.StatusOK {
				if feedBytes, err := io.ReadAll(respDisc.Body); err == nil && len(feedBytes) > 0 {
					var feed atomFeed
					if err := xml.Unmarshal(feedBytes, &feed); err == nil && len(feed.Entries) > 0 {
						items := a.parseAtomEntries(feed.Entries)
						if len(items) > 0 {
							return a.finalizeAnnouncements(items, cacheFile, hashFile), nil
						}
					}
				}
			}
		}
	}

	// 2. Secondary Source: GitHub Releases Feed
	releasesURL := "https://api.github.com/repos/kuraigu/Vinta/releases?per_page=5"
	reqRel, err := http.NewRequest("GET", releasesURL, nil)
	if err == nil {
		reqRel.Header.Set("User-Agent", "Vinta-App")
		reqRel.Header.Set("Accept", "application/vnd.github.v3+json")
		if token := resolveUpdaterToken(); token != "" {
			reqRel.Header.Set("Authorization", "Bearer "+token)
		}
		if respRel, err := client.Do(reqRel); err == nil {
			defer func() { _ = respRel.Body.Close() }()
			if respRel.StatusCode == http.StatusOK {
				if bodyBytes, err := io.ReadAll(respRel.Body); err == nil && len(bodyBytes) > 0 {
					var rawReleases []struct {
						ID          int64  `json:"id"`
						Name        string `json:"name"`
						Body        string `json:"body"`
						HTMLURL     string `json:"html_url"`
						TagName     string `json:"tag_name"`
						PublishedAt string `json:"published_at"`
						Author      struct {
							Login     string `json:"login"`
							AvatarURL string `json:"avatar_url"`
						} `json:"author"`
					}
					if err := json.Unmarshal(bodyBytes, &rawReleases); err == nil && len(rawReleases) > 0 {
						var items []AnnouncementItem
						for _, r := range rawReleases {
							title := r.Name
							if title == "" {
								title = r.TagName
							}
							authorName := r.Author.Login
							if authorName == "" {
								authorName = "kuraigu"
							}
							avatar := r.Author.AvatarURL
							if avatar == "" {
								avatar = fmt.Sprintf("https://github.com/%s.png?size=64", authorName)
							}
							items = append(items, AnnouncementItem{
								ID:           fmt.Sprintf("%d", r.ID),
								Title:        title,
								Body:         r.Body,
								URL:          r.HTMLURL,
								PublishedAt:  formatPublishedDate(r.PublishedAt),
								AuthorName:   authorName,
								AuthorAvatar: avatar,
								IsUnread:     false,
							})
						}
						return a.finalizeAnnouncements(items, cacheFile, hashFile), nil
					}
				}
			}
		}
	}

	// 3. Fallback: Load from offline disk cache or bundled starter announcement
	return a.loadCachedAnnouncements(cacheFile, hashFile), nil
}

func (a *App) parseAtomEntries(entries []atomEntry) []AnnouncementItem {
	var items []AnnouncementItem
	htmlTagRegex := regexp.MustCompile(`<[^>]*>`)
	for _, e := range entries {
		title := strings.TrimSpace(e.Title)
		if title == "" {
			continue
		}
		author := strings.TrimSpace(e.Author.Name)
		if author == "" {
			author = "kuraigu"
		}
		avatar := fmt.Sprintf("https://github.com/%s.png?size=64", author)
		cleanBody := htmlTagRegex.ReplaceAllString(e.Content, "")
		cleanBody = strings.TrimSpace(cleanBody)

		items = append(items, AnnouncementItem{
			ID:           e.ID,
			Title:        title,
			Body:         cleanBody,
			URL:          e.Link.Href,
			PublishedAt:  formatPublishedDate(e.Published),
			AuthorName:   author,
			AuthorAvatar: avatar,
			IsUnread:     false,
		})
	}
	return items
}

func (a *App) finalizeAnnouncements(items []AnnouncementItem, cacheFile, hashFile string) AnnouncementsPayload {
	if len(items) == 0 {
		return AnnouncementsPayload{Items: []AnnouncementItem{}, HasUnread: false}
	}

	hasher := sha256.New()
	hasher.Write([]byte(items[0].ID + items[0].PublishedAt))
	currentHash := hex.EncodeToString(hasher.Sum(nil))

	var storedHash string
	if hashFile != "" {
		if data, err := os.ReadFile(hashFile); err == nil {
			storedHash = strings.TrimSpace(string(data))
		}
	}

	hasUnread := storedHash != currentHash
	if hasUnread && len(items) > 0 {
		items[0].IsUnread = true
	}

	if cacheFile != "" {
		if cacheBytes, err := json.Marshal(items); err == nil {
			_ = os.WriteFile(cacheFile, cacheBytes, 0600)
		}
	}

	return AnnouncementsPayload{
		Items:     items,
		HasUnread: hasUnread,
	}
}

// MarkAnnouncementsRead clears the unread announcement state by recording the active hash.
func (a *App) MarkAnnouncementsRead() error {
	cacheDir, err := resolveCacheDir()
	if err != nil {
		return err
	}

	cacheFile := filepath.Join(cacheDir, "announcements.json")
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil
	}

	var items []AnnouncementItem
	if err := json.Unmarshal(data, &items); err != nil || len(items) == 0 {
		return nil
	}

	hasher := sha256.New()
	hasher.Write([]byte(items[0].ID + items[0].PublishedAt))
	currentHash := hex.EncodeToString(hasher.Sum(nil))

	hashFile := filepath.Join(cacheDir, "announcement_hash.txt")
	return os.WriteFile(hashFile, []byte(currentHash), 0600)
}

func (a *App) loadCachedAnnouncements(cacheFile, hashFile string) AnnouncementsPayload {
	if cacheFile != "" {
		data, err := os.ReadFile(cacheFile)
		if err == nil {
			var items []AnnouncementItem
			if err := json.Unmarshal(data, &items); err == nil && len(items) > 0 {
				var storedHash string
				if hashFile != "" {
					if d, err := os.ReadFile(hashFile); err == nil {
						storedHash = strings.TrimSpace(string(d))
					}
				}
				hasUnread := false
				hasher := sha256.New()
				hasher.Write([]byte(items[0].ID + items[0].PublishedAt))
				cur := hex.EncodeToString(hasher.Sum(nil))
				hasUnread = storedHash != cur
				if hasUnread {
					items[0].IsUnread = true
				}
				return AnnouncementsPayload{
					Items:     items,
					HasUnread: hasUnread,
				}
			}
		}
	}

	// First-run offline fallback when no cache exists yet.
	starterItem := AnnouncementItem{
		ID:           "welcome-announcement",
		Title:        "Welcome to Vinta!",
		Body:         "Developer announcements, project roadmap updates, and release highlights will be delivered directly here.\n\nConnect to the internet to automatically sync the latest news from GitHub.",
		URL:          "https://github.com/kuraigu/Vinta",
		PublishedAt:  time.Now().Format("Jan 02, 2006"),
		AuthorName:   "Vinta Team",
		AuthorAvatar: "https://github.com/kuraigu.png?size=64",
		IsUnread:     false,
	}
	return AnnouncementsPayload{
		Items:     []AnnouncementItem{starterItem},
		HasUnread: false,
	}
}

func formatPublishedDate(raw string) string {
	raw = strings.TrimSpace(raw)
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return t.Format("Jan 02, 2006")
	}
	return raw
}
