// Package backend provides core application services and offline caching for Vinta.
package backend

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
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
	cleanVersion = strings.TrimLeft(cleanVersion, "vV")
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
	req.Header.Set("Accept", "application/vnd.github+json")
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
	return fmt.Sprintf(`## What's New in v%s

Thanks for trying out Vinta! Here are a few things included in this build:

- **Top Island Bar**: Expands on click or hover and tucks away cleanly when finished.
- **Home Dashboard**: Quick glance at current weather, calendar events, and system settings.
- **Media Controls**: Adjust audio volume and control active music or video playback.
- **Notes & Timers**: Lightweight scratchpad, stopwatch, and timers ready from the dock.

*Release notes update automatically whenever a new version is published.*`, version)
}

// GraphQL response structure for GitHub Discussions (official API for private/public discussions).
type graphQLDiscussionsResponse struct {
	Data struct {
		Repository struct {
			Discussions struct {
				Nodes []struct {
					ID        string `json:"id"`
					Number    int    `json:"number"`
					Title     string `json:"title"`
					Body      string `json:"body"`
					URL       string `json:"url"`
					Published string `json:"publishedAt"`
					Category  struct {
						Name string `json:"name"`
						Slug string `json:"slug"`
					} `json:"category"`
					Author struct {
						Login     string `json:"login"`
						AvatarURL string `json:"avatarUrl"`
					} `json:"author"`
				} `json:"nodes"`
			} `json:"discussions"`
		} `json:"repository"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

// REST API structure for GitHub Discussions.
type ghDiscussion struct {
	NodeID    string `json:"node_id"`
	Number    int    `json:"number"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	HTMLURL   string `json:"html_url"`
	CreatedAt string `json:"created_at"`
	Category  struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	} `json:"category"`
	User struct {
		Login     string `json:"login"`
		AvatarURL string `json:"avatar_url"`
	} `json:"user"`
}

// Atom XML structures for public GitHub Discussions Announcements feed fallback.
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

	client := &http.Client{Timeout: 8 * time.Second}
	token := resolveUpdaterToken()

	if token == "" {
		log.Println("[Announcements] No UPDATER_TOKEN detected (.env / env / ldflags). Private repo API calls will fail.")
	} else {
		log.Printf("[Announcements] Token detected (length: %d). Querying GitHub Discussions...", len(token))
	}

	// 1. Primary Source (when token exists): GitHub GraphQL API (official GitHub Discussions API)
	if token != "" {
		if items, err := a.fetchGraphQLDiscussions(client, token); err == nil && len(items) > 0 {
			log.Printf("[Announcements] Loaded %d announcement(s) via GraphQL API.", len(items))
			return a.finalizeAnnouncements(items, cacheFile, hashFile), nil
		} else if err != nil {
			log.Printf("[Announcements] GraphQL fetch failed: %v", err)
		}
	}

	// 2. Secondary Source: GitHub REST API
	apiURL := "https://api.github.com/repos/kuraigu/Vinta/discussions?per_page=20"
	if reqAPI, err := http.NewRequest("GET", apiURL, nil); err == nil {
		reqAPI.Header.Set("User-Agent", "Vinta-App")
		reqAPI.Header.Set("Accept", "application/vnd.github+json")
		reqAPI.Header.Set("X-GitHub-Api-Version", "2022-11-28")
		if token != "" {
			reqAPI.Header.Set("Authorization", "Bearer "+token)
		}
		if respAPI, err := client.Do(reqAPI); err == nil {
			bodyBytes, _ := io.ReadAll(respAPI.Body)
			_ = respAPI.Body.Close()
			if respAPI.StatusCode == http.StatusOK && len(bodyBytes) > 0 {
				var discussions []ghDiscussion
				if err := json.Unmarshal(bodyBytes, &discussions); err == nil && len(discussions) > 0 {
					items := a.parseAPIDiscussions(discussions)
					if len(items) > 0 {
						log.Printf("[Announcements] Loaded %d announcement(s) via REST API.", len(items))
						return a.finalizeAnnouncements(items, cacheFile, hashFile), nil
					}
				}
			} else {
				log.Printf("[Announcements] REST API returned HTTP %d: %s", respAPI.StatusCode, string(bodyBytes))
			}
		}
	}

	// 3. Tertiary Source: Public Atom Feed (works on public repos without a token)
	discussionsURL := "https://github.com/kuraigu/Vinta/discussions/categories/announcements.atom"
	if reqDisc, err := http.NewRequest("GET", discussionsURL, nil); err == nil {
		reqDisc.Header.Set("User-Agent", "Vinta-App")
		if respDisc, err := client.Do(reqDisc); err == nil {
			feedBytes, _ := io.ReadAll(respDisc.Body)
			_ = respDisc.Body.Close()
			if respDisc.StatusCode == http.StatusOK && len(feedBytes) > 0 {
				var feed atomFeed
				if err := xml.Unmarshal(feedBytes, &feed); err == nil && len(feed.Entries) > 0 {
					items := a.parseAtomEntries(feed.Entries)
					if len(items) > 0 {
						log.Printf("[Announcements] Loaded %d announcement(s) via Atom feed.", len(items))
						return a.finalizeAnnouncements(items, cacheFile, hashFile), nil
					}
				}
			}
		}
	}

	// 4. Fallback: Load from offline disk cache or starter welcome announcement
	log.Println("[Announcements] All remote sources failed or returned 0 announcements. Using fallback.")
	return a.loadCachedAnnouncements(cacheFile, hashFile), nil
}

func (a *App) fetchGraphQLDiscussions(client *http.Client, token string) ([]AnnouncementItem, error) {
	query := `query {
		repository(owner: "kuraigu", name: "Vinta") {
			discussions(first: 20, orderBy: {field: CREATED_AT, direction: DESC}) {
				nodes {
					id
					number
					title
					body
					url
					publishedAt
					category {
						name
						slug
					}
					author {
						login
						avatarUrl
					}
				}
			}
		}
	}`

	payloadBytes, err := json.Marshal(map[string]string{"query": query})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://api.github.com/graphql", bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Vinta-App")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var gqlResp graphQLDiscussionsResponse
	if err := json.Unmarshal(bodyBytes, &gqlResp); err != nil {
		return nil, err
	}
	if len(gqlResp.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL error: %s", gqlResp.Errors[0].Message)
	}

	nodes := gqlResp.Data.Repository.Discussions.Nodes
	var allItems []AnnouncementItem
	var announcementItems []AnnouncementItem

	for _, n := range nodes {
		title := strings.TrimSpace(n.Title)
		if title == "" {
			continue
		}
		author := strings.TrimSpace(n.Author.Login)
		if author == "" {
			author = "kuraigu"
		}
		avatar := strings.TrimSpace(n.Author.AvatarURL)
		if avatar == "" {
			avatar = fmt.Sprintf("https://github.com/%s.png?size=64", author)
		}

		item := AnnouncementItem{
			ID:           n.ID,
			Title:        title,
			Body:         strings.TrimSpace(n.Body),
			URL:          n.URL,
			PublishedAt:  formatPublishedDate(n.Published),
			AuthorName:   author,
			AuthorAvatar: avatar,
			IsUnread:     false,
		}

		allItems = append(allItems, item)

		slug := strings.ToLower(strings.TrimSpace(n.Category.Slug))
		catName := strings.ToLower(strings.TrimSpace(n.Category.Name))
		if strings.Contains(slug, "announcement") || strings.Contains(catName, "announcement") {
			announcementItems = append(announcementItems, item)
		}
	}

	// Prefer discussions in the Announcements category; if posted in General/another category, still return them
	if len(announcementItems) > 0 {
		return announcementItems, nil
	}
	return allItems, nil
}

func (a *App) parseAPIDiscussions(discussions []ghDiscussion) []AnnouncementItem {
	var allItems []AnnouncementItem
	var announcementItems []AnnouncementItem

	for _, d := range discussions {
		title := strings.TrimSpace(d.Title)
		if title == "" {
			continue
		}

		author := strings.TrimSpace(d.User.Login)
		if author == "" {
			author = "kuraigu"
		}
		avatar := strings.TrimSpace(d.User.AvatarURL)
		if avatar == "" {
			avatar = fmt.Sprintf("https://github.com/%s.png?size=64", author)
		}

		id := d.NodeID
		if id == "" {
			id = fmt.Sprintf("discussion-%d", d.Number)
		}

		item := AnnouncementItem{
			ID:           id,
			Title:        title,
			Body:         strings.TrimSpace(d.Body),
			URL:          d.HTMLURL,
			PublishedAt:  formatPublishedDate(d.CreatedAt),
			AuthorName:   author,
			AuthorAvatar: avatar,
			IsUnread:     false,
		}

		allItems = append(allItems, item)

		slug := strings.ToLower(strings.TrimSpace(d.Category.Slug))
		catName := strings.ToLower(strings.TrimSpace(d.Category.Name))
		if strings.Contains(slug, "announcement") || strings.Contains(catName, "announcement") {
			announcementItems = append(announcementItems, item)
		}
	}

	if len(announcementItems) > 0 {
		return announcementItems
	}
	return allItems
}

func (a *App) parseAtomEntries(entries []atomEntry) []AnnouncementItem {
	var items []AnnouncementItem
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
		cleanBody := cleanDiscussionContent(e.Content)

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

// cleanDiscussionContent normalizes HTML from the Atom feed into clean Markdown, preserving images and links.
func cleanDiscussionContent(content string) string {
	if content == "" {
		return ""
	}

	// 1. Convert <img> tags to markdown image syntax: ![alt](src)
	imgRegex := regexp.MustCompile(`(?i)<img[^>]*src=["']([^"']+)["'][^>]*alt=["']([^"']*)["'][^>]*>`)
	content = imgRegex.ReplaceAllString(content, "![$2]($1)")

	imgFallbackRegex := regexp.MustCompile(`(?i)<img[^>]*src=["']([^"']+)["'][^>]*>`)
	content = imgFallbackRegex.ReplaceAllString(content, "![]($1)")

	// 2. Convert <a> tags to markdown link syntax: [text](href)
	linkRegex := regexp.MustCompile(`(?i)<a[^>]*href=["']([^"']+)["'][^>]*>(.*?)</a>`)
	content = linkRegex.ReplaceAllString(content, "[$2]($1)")

	// 3. Convert paragraphs & breaks to newlines
	brRegex := regexp.MustCompile(`(?i)<br\s*/?>`)
	content = brRegex.ReplaceAllString(content, "\n")

	pCloseRegex := regexp.MustCompile(`(?i)</p>`)
	content = pCloseRegex.ReplaceAllString(content, "\n\n")

	// 4. Convert basic styling tags
	bRegex := regexp.MustCompile(`(?i)<(?:strong|b)>(.*?)</(?:strong|b)>`)
	content = bRegex.ReplaceAllString(content, "**$1**")

	iRegex := regexp.MustCompile(`(?i)<(?:em|i)>(.*?)</(?:em|i)>`)
	content = iRegex.ReplaceAllString(content, "*$1*")

	liRegex := regexp.MustCompile(`(?i)<li[^>]*>(.*?)</li>`)
	content = liRegex.ReplaceAllString(content, "* $1\n")

	// 5. Strip any remaining unsupported/dangerous HTML tags
	otherTagsRegex := regexp.MustCompile(`</?[a-zA-Z][^>]*>`)
	content = otherTagsRegex.ReplaceAllString(content, "")

	// 6. Decode common HTML entities
	content = strings.ReplaceAll(content, "&amp;", "&")
	content = strings.ReplaceAll(content, "&lt;", "<")
	content = strings.ReplaceAll(content, "&gt;", ">")
	content = strings.ReplaceAll(content, "&quot;", "\"")
	content = strings.ReplaceAll(content, "&#39;", "'")
	content = strings.ReplaceAll(content, "&nbsp;", " ")

	return strings.TrimSpace(content)
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
				if strings.Contains(items[0].URL, "/releases/") || strings.HasPrefix(items[0].Title, "Release ") || items[0].ID == "welcome-announcement" {
					_ = os.Remove(cacheFile)
					if hashFile != "" {
						_ = os.Remove(hashFile)
					}
					return a.starterAnnouncement()
				}

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

	return a.starterAnnouncement()
}

func (a *App) starterAnnouncement() AnnouncementsPayload {
	starterItem := AnnouncementItem{
		ID:           "welcome-announcement",
		Title:        "Welcome to Vinta!",
		Body:         "Thanks for installing Vinta. Official announcements, roadmap milestones, and community news will appear here once published.\n\nCheck out the documentation or join our GitHub Discussions to connect with the team.",
		URL:          "https://github.com/kuraigu/Vinta/discussions",
		PublishedAt:  time.Now().Format("Jan 02, 2006"),
		AuthorName:   "kuraigu",
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