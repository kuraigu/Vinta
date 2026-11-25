// Package backend provides core application services and state management for Vinta.
package backend

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// GitHubStatus reports the current GitHub authentication state.
type GitHubStatus struct {
	Connected        bool   `json:"connected"`
	Username         string `json:"username"`
	AvatarURL        string `json:"avatarUrl"`
	HasOAuthClientID bool   `json:"hasOAuthClientId"`
}

// GitHubDeviceCodeInfo contains the user code and verification URL for GitHub OAuth Device Flow.
type GitHubDeviceCodeInfo struct {
	DeviceCode      string `json:"deviceCode"`
	UserCode        string `json:"userCode"`
	VerificationURI string `json:"verificationUri"`
	ExpiresIn       int    `json:"expiresIn"`
	Interval        int    `json:"interval"`
}

// GitHubContributionDay represents a single day cell in the contribution heatmap.
type GitHubContributionDay struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
	Level int    `json:"level"`
}

// GitHubItem represents a GitHub PR, issue, or notification item.
type GitHubItem struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Repo      string `json:"repo"`
	URL       string `json:"url"`
	Kind      string `json:"kind"`
	UpdatedAt string `json:"updatedAt"`
	State     string `json:"state"`
}

// GitHubDashboard aggregates developer activity for the Developer Hub page and minimal pill.
type GitHubDashboard struct {
	Username       string                  `json:"username"`
	AvatarURL      string                  `json:"avatarUrl"`
	CurrentStreak  int                     `json:"currentStreak"`
	TotalRecent    int                     `json:"totalRecent"`
	UnreadCount    int                     `json:"unreadCount"`
	Contributions  []GitHubContributionDay `json:"contributions"`
	Notifications  []GitHubItem            `json:"notifications"`
	PullRequests   []GitHubItem            `json:"pullRequests"`
	AssignedIssues []GitHubItem            `json:"assignedIssues"`
	FetchedAt      int64                   `json:"fetchedAt"`
}

// AIProviderUsage reports configuration and health status for an AI coding/model provider.
type AIProviderUsage struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Configured   bool   `json:"configured"`
	Status       string `json:"status"`
	Detail       string `json:"detail"`
	DashboardURL string `json:"dashboardUrl"`
}

// GetGitHubStatus returns whether GitHub is connected in the DPAPI vault.
func (a *App) GetGitHubStatus() GitHubStatus {
	vault, _ := readCredentialVault()
	return GitHubStatus{
		Connected:        strings.TrimSpace(vault.GitHub.AccessToken) != "",
		Username:         vault.GitHub.Username,
		AvatarURL:        vault.GitHub.AvatarURL,
		HasOAuthClientID: GetEnvConfig("VINTA_GITHUB_CLIENT_ID") != "",
	}
}

// ConnectGitHubWithPAT validates a GitHub Personal Access Token and stores it in DPAPI.
func (a *App) ConnectGitHubWithPAT(token string) (GitHubStatus, error) {
	trimmed := strings.TrimSpace(token)
	if trimmed == "" {
		return a.GetGitHubStatus(), fmt.Errorf("GitHub token cannot be empty")
	}
	username, avatar, err := fetchGitHubViewer(trimmed)
	if err != nil {
		return a.GetGitHubStatus(), err
	}
	if err := updateCredentialVault(func(v *credentialVault) {
		v.GitHub = githubCredentialRecord{
			AccessToken: trimmed,
			Username:    username,
			AvatarURL:   avatar,
			ConnectedAt: time.Now().Unix(),
		}
	}); err != nil {
		return a.GetGitHubStatus(), err
	}
	return a.GetGitHubStatus(), nil
}

// StartGitHubDeviceFlow initiates GitHub OAuth Device Flow and opens the verification page.
func (a *App) StartGitHubDeviceFlow() (GitHubDeviceCodeInfo, error) {
	clientID := GetEnvConfig("VINTA_GITHUB_CLIENT_ID")
	if clientID == "" {
		return GitHubDeviceCodeInfo{}, fmt.Errorf("github sign-in is not configured in this build")
	}
	form := url.Values{}
	form.Set("client_id", clientID)
	// GitHub Apps (Iv... Client IDs) use fine-grained read-only permissions configured on the app itself.
	// Classic OAuth Apps (Ov...) use least-privilege read:user + notifications scope.
	if !strings.HasPrefix(clientID, "Iv") {
		form.Set("scope", "read:user notifications")
	}

	req, err := http.NewRequest(http.MethodPost, "https://github.com/login/device/code", strings.NewReader(form.Encode()))
	if err != nil {
		return GitHubDeviceCodeInfo{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := sharedAPIClient.Do(req)
	if err != nil {
		return GitHubDeviceCodeInfo{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return GitHubDeviceCodeInfo{}, fmt.Errorf("github device flow returned status %d", resp.StatusCode)
	}

	var parsed struct {
		DeviceCode      string `json:"device_code"`
		UserCode        string `json:"user_code"`
		VerificationURI string `json:"verification_uri"`
		ExpiresIn       int    `json:"expires_in"`
		Interval        int    `json:"interval"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return GitHubDeviceCodeInfo{}, err
	}
	if parsed.VerificationURI != "" {
		_ = a.OpenURL(parsed.VerificationURI)
	}
	if parsed.Interval < 5 {
		parsed.Interval = 5
	}
	return GitHubDeviceCodeInfo{
		DeviceCode:      parsed.DeviceCode,
		UserCode:        parsed.UserCode,
		VerificationURI: parsed.VerificationURI,
		ExpiresIn:       parsed.ExpiresIn,
		Interval:        parsed.Interval,
	}, nil
}

// PollGitHubDeviceFlow polls GitHub until the user completes authorization in the browser.
func (a *App) PollGitHubDeviceFlow(deviceCode string, interval int) (GitHubStatus, error) {
	clientID := GetEnvConfig("VINTA_GITHUB_CLIENT_ID")
	if clientID == "" || strings.TrimSpace(deviceCode) == "" {
		return a.GetGitHubStatus(), fmt.Errorf("invalid GitHub device flow state")
	}
	if interval < 5 {
		interval = 5
	}
	deadline := time.Now().Add(3 * time.Minute)
	for time.Now().Before(deadline) {
		time.Sleep(time.Duration(interval) * time.Second)
		form := url.Values{}
		form.Set("client_id", clientID)
		form.Set("device_code", deviceCode)
		form.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")

		req, err := http.NewRequest(http.MethodPost, "https://github.com/login/oauth/access_token", strings.NewReader(form.Encode()))
		if err != nil {
			return a.GetGitHubStatus(), err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Accept", "application/json")

		resp, err := sharedAPIClient.Do(req)
		if err != nil {
			continue
		}
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
		_ = resp.Body.Close()

		var res struct {
			AccessToken string `json:"access_token"`
			Error       string `json:"error"`
		}
		if err := json.Unmarshal(body, &res); err != nil {
			continue
		}
		if res.AccessToken != "" {
			return a.ConnectGitHubWithPAT(res.AccessToken)
		}
		switch res.Error {
		case "authorization_pending":
			continue
		case "slow_down":
			interval += 5
			continue
		case "expired_token", "access_denied":
			return a.GetGitHubStatus(), fmt.Errorf("github authorization %s", res.Error)
		}
	}
	return a.GetGitHubStatus(), fmt.Errorf("github device sign-in timed out")
}

// DisconnectGitHub removes the stored GitHub token from the DPAPI vault.
func (a *App) DisconnectGitHub() error {
	return updateCredentialVault(func(v *credentialVault) {
		v.GitHub = githubCredentialRecord{}
	})
}

func fetchGitHubViewer(token string) (username, avatar string, err error) {
	req, err := http.NewRequest(http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "Vinta-Desktop")
	resp, err := sharedAPIClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("invalid GitHub token (status %d)", resp.StatusCode)
	}
	var user struct {
		Login     string `json:"login"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return "", "", err
	}
	return user.Login, user.AvatarURL, nil
}

// GetGitHubDashboard fetches cached contributions, notifications, PRs, and assigned issues.
func (a *App) GetGitHubDashboard(forceRefresh bool) (GitHubDashboard, error) {
	vault, err := readCredentialVault()
	if err != nil || strings.TrimSpace(vault.GitHub.AccessToken) == "" {
		return GitHubDashboard{}, fmt.Errorf("GitHub is not connected")
	}
	token := vault.GitHub.AccessToken
	username := vault.GitHub.Username
	ctx := context.Background()

	dash := GitHubDashboard{
		Username:       username,
		AvatarURL:      vault.GitHub.AvatarURL,
		Contributions:  []GitHubContributionDay{},
		Notifications:  []GitHubItem{},
		PullRequests:   []GitHubItem{},
		AssignedIssues: []GitHubItem{},
		FetchedAt:      time.Now().Unix(),
	}

	// 1. Contributions via GraphQL (cached 10m)
	gqlQuery := map[string]any{
		"query": `query { viewer { contributionsCollection { contributionCalendar { totalContributions weeks { contributionDays { date contributionCount } } } } } }`,
	}
	gqlBytes, _ := json.Marshal(gqlQuery)
	if body, gqlErr := cachedHTTPDo(ctx, "gh_contrib_"+username, 10*time.Minute, forceRefresh, func(c context.Context) (*http.Request, error) {
		req, reqErr := http.NewRequestWithContext(c, http.MethodPost, "https://api.github.com/graphql", bytes.NewReader(gqlBytes))
		if reqErr != nil {
			return nil, reqErr
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "Vinta-Desktop")
		return req, nil
	}); gqlErr == nil {
		var gqlRes struct {
			Data struct {
				Viewer struct {
					ContributionsCollection struct {
						ContributionCalendar struct {
							TotalContributions int `json:"totalContributions"`
							Weeks              []struct {
								ContributionDays []struct {
									Date              string `json:"date"`
									ContributionCount int    `json:"contributionCount"`
								} `json:"contributionDays"`
							} `json:"weeks"`
						} `json:"contributionCalendar"`
					} `json:"contributionsCollection"`
				} `json:"viewer"`
			} `json:"data"`
		}
		if json.Unmarshal(body, &gqlRes) == nil {
			cal := gqlRes.Data.Viewer.ContributionsCollection.ContributionCalendar
			dash.TotalRecent = cal.TotalContributions
			allDays := make([]GitHubContributionDay, 0, 370)
			for _, w := range cal.Weeks {
				for _, d := range w.ContributionDays {
					lvl := 0
					switch {
					case d.ContributionCount >= 10:
						lvl = 4
					case d.ContributionCount >= 6:
						lvl = 3
					case d.ContributionCount >= 3:
						lvl = 2
					case d.ContributionCount >= 1:
						lvl = 1
					}
					allDays = append(allDays, GitHubContributionDay{
						Date:  d.Date,
						Count: d.ContributionCount,
						Level: lvl,
					})
				}
			}
			// Compute current streak counting backwards from today (or yesterday if today is 0).
			streak := 0
			for i := len(allDays) - 1; i >= 0; i-- {
				if allDays[i].Count > 0 {
					streak++
				} else if i == len(allDays)-1 {
					continue
				} else {
					break
				}
			}
			dash.CurrentStreak = streak
			// Keep the latest 84 days (12 weeks) for the compact heatmap.
			if len(allDays) > 84 {
				dash.Contributions = allDays[len(allDays)-84:]
			} else {
				dash.Contributions = allDays
			}
		}
	}

	// 2. Unread notifications (cached 3m)
	if body, nErr := cachedHTTPDo(ctx, "gh_notifs_"+username, 3*time.Minute, forceRefresh, func(c context.Context) (*http.Request, error) {
		req, reqErr := http.NewRequestWithContext(c, http.MethodGet, "https://api.github.com/notifications?per_page=10", nil)
		if reqErr != nil {
			return nil, reqErr
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept", "application/vnd.github+json")
		req.Header.Set("User-Agent", "Vinta-Desktop")
		return req, nil
	}); nErr == nil {
		var rawNotifs []struct {
			ID         string `json:"id"`
			UpdatedAt  string `json:"updated_at"`
			Repository struct {
				FullName string `json:"full_name"`
				HTMLURL  string `json:"html_url"`
			} `json:"repository"`
			Subject struct {
				Title string `json:"title"`
				Type  string `json:"type"`
			} `json:"subject"`
		}
		if json.Unmarshal(body, &rawNotifs) == nil {
			dash.UnreadCount = len(rawNotifs)
			for _, n := range rawNotifs {
				dash.Notifications = append(dash.Notifications, GitHubItem{
					ID:        n.ID,
					Title:     n.Subject.Title,
					Repo:      n.Repository.FullName,
					URL:       n.Repository.HTMLURL,
					Kind:      n.Subject.Type,
					UpdatedAt: n.UpdatedAt,
				})
			}
		}
	}

	// 3. Open PRs involving the user (cached 5m)
	dash.PullRequests = fetchGitHubSearchItems(ctx, token, "gh_prs_"+username, "is:pr+is:open+involves:"+url.QueryEscape(username), "PR", forceRefresh)

	// 4. Assigned open issues (cached 5m)
	dash.AssignedIssues = fetchGitHubSearchItems(ctx, token, "gh_issues_"+username, "is:issue+is:open+assignee:"+url.QueryEscape(username), "Issue", forceRefresh)

	return dash, nil
}

func fetchGitHubSearchItems(ctx context.Context, token, cacheKey, query, kind string, forceRefresh bool) []GitHubItem {
	endpoint := "https://api.github.com/search/issues?q=" + query + "&sort=updated&per_page=8"
	body, err := cachedHTTPDo(ctx, cacheKey, 5*time.Minute, forceRefresh, func(c context.Context) (*http.Request, error) {
		req, reqErr := http.NewRequestWithContext(c, http.MethodGet, endpoint, nil)
		if reqErr != nil {
			return nil, reqErr
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept", "application/vnd.github+json")
		req.Header.Set("User-Agent", "Vinta-Desktop")
		return req, nil
	})
	if err != nil {
		return []GitHubItem{}
	}
	var parsed struct {
		Items []struct {
			ID            int64  `json:"id"`
			Title         string `json:"title"`
			HTMLURL       string `json:"html_url"`
			RepositoryURL string `json:"repository_url"`
			UpdatedAt     string `json:"updated_at"`
			State         string `json:"state"`
		} `json:"items"`
	}
	if json.Unmarshal(body, &parsed) != nil {
		return []GitHubItem{}
	}
	items := make([]GitHubItem, 0, len(parsed.Items))
	for _, it := range parsed.Items {
		repo := strings.TrimPrefix(it.RepositoryURL, "https://api.github.com/repos/")
		items = append(items, GitHubItem{
			ID:        fmt.Sprintf("%d", it.ID),
			Title:     it.Title,
			Repo:      repo,
			URL:       it.HTMLURL,
			Kind:      kind,
			UpdatedAt: it.UpdatedAt,
			State:     it.State,
		})
	}
	return items
}

// GetAIUsageOverview returns monthly spend (for Admin keys) or verified model count for Claude and OpenAI.
func (a *App) GetAIUsageOverview(forceRefresh bool) []AIProviderUsage {
	vault, _ := readCredentialVault()
	ctx := context.Background()
	now := time.Now().UTC()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	claude := AIProviderUsage{
		ID:           "claude",
		Name:         "Claude",
		Configured:   strings.TrimSpace(vault.AIKeys.ClaudeKey) != "",
		Status:       "No key",
		Detail:       "console.anthropic.com",
		DashboardURL: "https://console.anthropic.com/settings/usage",
	}
	if claude.Configured {
		costURL := fmt.Sprintf(
			"https://api.anthropic.com/v1/organizations/cost_report?starting_at=%s&ending_at=%s&bucket_width=1d",
			url.QueryEscape(monthStart.Format(time.RFC3339)),
			url.QueryEscape(now.Format(time.RFC3339)),
		)
		if body, err := cachedHTTPDo(ctx, "ai_cost_claude", 15*time.Minute, forceRefresh, func(c context.Context) (*http.Request, error) {
			req, rErr := http.NewRequestWithContext(c, http.MethodGet, costURL, nil)
			if rErr != nil {
				return nil, rErr
			}
			req.Header.Set("x-api-key", vault.AIKeys.ClaudeKey)
			req.Header.Set("anthropic-version", "2023-06-01")
			return req, nil
		}); err == nil {
			var report struct {
				Data []struct {
					Results []struct {
						Amount string `json:"amount"`
					} `json:"results"`
				} `json:"data"`
			}
			if json.Unmarshal(body, &report) == nil {
				var totalCents float64
				for _, bucket := range report.Data {
					for _, r := range bucket.Results {
						var val float64
						if _, scanErr := fmt.Sscanf(r.Amount, "%f", &val); scanErr == nil {
							totalCents += val
						}
					}
				}
				claude.Status = fmt.Sprintf("$%.2f this mo", totalCents/100.0)
				claude.Detail = "console.anthropic.com · Admin key"
			}
		} else if body, mErr := cachedHTTPDo(ctx, "ai_models_claude", 15*time.Minute, forceRefresh, func(c context.Context) (*http.Request, error) {
			req, rErr := http.NewRequestWithContext(c, http.MethodGet, "https://api.anthropic.com/v1/models?limit=100", nil)
			if rErr != nil {
				return nil, rErr
			}
			req.Header.Set("x-api-key", vault.AIKeys.ClaudeKey)
			req.Header.Set("anthropic-version", "2023-06-01")
			return req, nil
		}); mErr == nil {
			var modelsRes struct {
				Data []struct {
					ID string `json:"id"`
				} `json:"data"`
			}
			if json.Unmarshal(body, &modelsRes) == nil && len(modelsRes.Data) > 0 {
				claude.Status = fmt.Sprintf("Connected (%d models)", len(modelsRes.Data))
			} else {
				claude.Status = "Connected"
			}
			claude.Detail = "Standard key · use Admin key for spend"
		} else {
			claude.Status = "Invalid key"
		}
	}

	openai := AIProviderUsage{
		ID:           "openai",
		Name:         "OpenAI",
		Configured:   strings.TrimSpace(vault.AIKeys.OpenAIKey) != "",
		Status:       "No key",
		Detail:       "platform.openai.com",
		DashboardURL: "https://platform.openai.com/usage",
	}
	if openai.Configured {
		costURL := fmt.Sprintf("https://api.openai.com/v1/organization/costs?start_time=%d&limit=31", monthStart.Unix())
		if body, err := cachedHTTPDo(ctx, "ai_cost_openai", 15*time.Minute, forceRefresh, func(c context.Context) (*http.Request, error) {
			req, rErr := http.NewRequestWithContext(c, http.MethodGet, costURL, nil)
			if rErr != nil {
				return nil, rErr
			}
			req.Header.Set("Authorization", "Bearer "+vault.AIKeys.OpenAIKey)
			return req, nil
		}); err == nil {
			var report struct {
				Data []struct {
					Results []struct {
						Amount struct {
							Value float64 `json:"value"`
						} `json:"amount"`
					} `json:"results"`
				} `json:"data"`
			}
			if json.Unmarshal(body, &report) == nil {
				var totalUSD float64
				for _, bucket := range report.Data {
					for _, r := range bucket.Results {
						totalUSD += r.Amount.Value
					}
				}
				openai.Status = fmt.Sprintf("$%.2f this mo", totalUSD)
				openai.Detail = "platform.openai.com · Admin key"
			}
		} else if body, mErr := cachedHTTPDo(ctx, "ai_models_openai", 15*time.Minute, forceRefresh, func(c context.Context) (*http.Request, error) {
			req, rErr := http.NewRequestWithContext(c, http.MethodGet, "https://api.openai.com/v1/models", nil)
			if rErr != nil {
				return nil, rErr
			}
			req.Header.Set("Authorization", "Bearer "+vault.AIKeys.OpenAIKey)
			return req, nil
		}); mErr == nil {
			var modelsRes struct {
				Data []struct {
					ID string `json:"id"`
				} `json:"data"`
			}
			if json.Unmarshal(body, &modelsRes) == nil && len(modelsRes.Data) > 0 {
				openai.Status = fmt.Sprintf("Connected (%d models)", len(modelsRes.Data))
			} else {
				openai.Status = "Connected"
			}
			openai.Detail = "Standard key · use Admin key for spend"
		} else {
			openai.Status = "Invalid key"
		}
	}

	return []AIProviderUsage{claude, openai}
}
