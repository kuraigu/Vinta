// Package backend provides core application services and state management for Vinta.
package backend

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// GoogleCalendarStatus describes the current Google Calendar OAuth connection state.
type GoogleCalendarStatus struct {
	Connected    bool   `json:"connected"`
	Email        string `json:"email"`
	HasClientID  bool   `json:"hasClientId"`
	LastSyncedAt int64  `json:"lastSyncedAt"`
}

// GoogleCalendarEventDTO represents a normalized event fetched from Google Calendar API v3.
type GoogleCalendarEventDTO struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Date        string `json:"date"`
	StartTime   string `json:"startTime"`
	EndTime     string `json:"endTime"`
	IsAllDay    bool   `json:"isAllDay"`
	Location    string `json:"location"`
	HangoutLink string `json:"hangoutLink"`
	IsHoliday   bool   `json:"isHoliday"`
	CalendarID  string `json:"calendarId"`
}

type googleTokenVault struct {
	ClientID     string `json:"clientId"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresAt    int64  `json:"expiresAt"`
	Email        string `json:"email"`
	LastSyncedAt int64  `json:"lastSyncedAt"`
}

var dpapiGCalHeader = []byte("VNGC")

func getGoogleTokenPath() string {
	return filepath.Join(GetAppDataDir(), "gcal_tokens.bin")
}

func loadGoogleTokenVault() (*googleTokenVault, error) {
	raw, err := os.ReadFile(getGoogleTokenPath())
	if err != nil || len(raw) == 0 {
		return &googleTokenVault{}, nil
	}
	if len(raw) >= len(dpapiGCalHeader) && string(raw[:len(dpapiGCalHeader)]) == string(dpapiGCalHeader) {
		decrypted, decErr := decryptDPAPI(raw[len(dpapiGCalHeader):])
		if decErr != nil {
			return &googleTokenVault{}, decErr
		}
		raw = decrypted
	}
	var vault googleTokenVault
	if err := json.Unmarshal(raw, &vault); err != nil {
		return &googleTokenVault{}, err
	}
	return &vault, nil
}

func saveGoogleTokenVault(vault *googleTokenVault) error {
	dir := GetAppDataDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	plain, err := json.Marshal(vault)
	if err != nil {
		return err
	}
	encrypted, err := encryptDPAPI(plain)
	if err != nil {
		return err
	}
	out := append(append([]byte{}, dpapiGCalHeader...), encrypted...)
	return os.WriteFile(getGoogleTokenPath(), out, 0600)
}

func resolveGoogleClientID(override string) string {
	if trimmed := strings.TrimSpace(override); trimmed != "" {
		return trimmed
	}
	if env := strings.TrimSpace(os.Getenv("VINTA_GOOGLE_CLIENT_ID")); env != "" {
		return env
	}
	if vault, err := loadGoogleTokenVault(); err == nil && strings.TrimSpace(vault.ClientID) != "" {
		return strings.TrimSpace(vault.ClientID)
	}
	return ""
}

// GetGoogleCalendarStatus returns whether Google Calendar is connected via DPAPI tokens.
func (a *App) GetGoogleCalendarStatus() GoogleCalendarStatus {
	vault, _ := loadGoogleTokenVault()
	clientID := resolveGoogleClientID("")
	return GoogleCalendarStatus{
		Connected:    strings.TrimSpace(vault.RefreshToken) != "" || strings.TrimSpace(vault.AccessToken) != "",
		Email:        vault.Email,
		HasClientID:  clientID != "",
		LastSyncedAt: vault.LastSyncedAt,
	}
}

// DisconnectGoogleCalendar removes the DPAPI-encrypted Google Calendar token vault.
func (a *App) DisconnectGoogleCalendar() error {
	err := os.Remove(getGoogleTokenPath())
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func generatePKCEPair() (verifier string, challenge string, err error) {
	buf := make([]byte, 32)
	if _, err = rand.Read(buf); err != nil {
		return "", "", err
	}
	verifier = base64.RawURLEncoding.EncodeToString(buf)
	sum := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(sum[:])
	return verifier, challenge, nil
}

// ConnectGoogleCalendar initiates an RFC 8252 loopback OAuth 2.0 PKCE flow in the system browser.
func (a *App) ConnectGoogleCalendar(clientID string) (GoogleCalendarStatus, error) {
	resolvedID := resolveGoogleClientID(clientID)
	if resolvedID == "" {
		return a.GetGoogleCalendarStatus(), fmt.Errorf("missing Google OAuth Client ID (set VINTA_GOOGLE_CLIENT_ID or enter a Desktop OAuth Client ID)")
	}

	verifier, challenge, err := generatePKCEPair()
	if err != nil {
		return a.GetGoogleCalendarStatus(), err
	}

	stateBuf := make([]byte, 16)
	if _, err := rand.Read(stateBuf); err != nil {
		return a.GetGoogleCalendarStatus(), err
	}
	oauthState := base64.RawURLEncoding.EncodeToString(stateBuf)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return a.GetGoogleCalendarStatus(), fmt.Errorf("failed to start loopback listener: %w", err)
	}
	redirectURI := fmt.Sprintf("http://%s/callback", ln.Addr().String())

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("state") != oauthState {
			http.Error(w, "Invalid OAuth state", http.StatusBadRequest)
			errCh <- fmt.Errorf("oauth state mismatch")
			return
		}
		if oauthErr := q.Get("error"); oauthErr != "" {
			http.Error(w, "Authorization cancelled: "+oauthErr, http.StatusBadRequest)
			errCh <- fmt.Errorf("oauth error: %s", oauthErr)
			return
		}
		code := q.Get("code")
		if code == "" {
			http.Error(w, "Missing authorization code", http.StatusBadRequest)
			errCh <- fmt.Errorf("missing authorization code")
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte("<html><body style='font-family:sans-serif;text-align:center;padding:40px'><h2>Google Calendar connected to Vinta</h2><p>You can close this tab and return to Vinta.</p></body></html>"))
		codeCh <- code
	})

	srv := &http.Server{Handler: mux}
	go func() {
		_ = srv.Serve(ln)
	}()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}()

	authURL := "https://accounts.google.com/o/oauth2/v2/auth?" + url.Values{
		"client_id":             {resolvedID},
		"redirect_uri":          {redirectURI},
		"response_type":         {"code"},
		"scope":                 {"https://www.googleapis.com/auth/calendar.readonly email"},
		"code_challenge":        {challenge},
		"code_challenge_method": {"S256"},
		"access_type":           {"offline"},
		"prompt":                {"consent"},
		"state":                 {oauthState},
	}.Encode()

	if openErr := a.OpenURL(authURL); openErr != nil {
		return a.GetGoogleCalendarStatus(), openErr
	}

	var authCode string
	select {
	case authCode = <-codeCh:
	case oauthErr := <-errCh:
		return a.GetGoogleCalendarStatus(), oauthErr
	case <-time.After(3 * time.Minute):
		return a.GetGoogleCalendarStatus(), fmt.Errorf("google authorization timed out")
	}

	tokenForm := url.Values{
		"client_id":     {resolvedID},
		"code":          {authCode},
		"code_verifier": {verifier},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {redirectURI},
	}

	httpClient := &http.Client{Timeout: 15 * time.Second}
	resp, err := httpClient.PostForm("https://oauth2.googleapis.com/token", tokenForm)
	if err != nil {
		return a.GetGoogleCalendarStatus(), fmt.Errorf("token exchange failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	if resp.StatusCode != http.StatusOK {
		return a.GetGoogleCalendarStatus(), fmt.Errorf("token exchange returned %d", resp.StatusCode)
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return a.GetGoogleCalendarStatus(), err
	}

	email := fetchGoogleUserEmail(httpClient, tokenResp.AccessToken)
	vault := &googleTokenVault{
		ClientID:     resolvedID,
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresAt:    time.Now().Unix() + tokenResp.ExpiresIn - 60,
		Email:        email,
		LastSyncedAt: time.Now().UnixMilli(),
	}
	if err := saveGoogleTokenVault(vault); err != nil {
		return a.GetGoogleCalendarStatus(), err
	}

	return a.GetGoogleCalendarStatus(), nil
}

func fetchGoogleUserEmail(client *http.Client, accessToken string) string {
	req, err := http.NewRequest(http.MethodGet, "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return ""
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			_ = resp.Body.Close()
		}
		return ""
	}
	defer func() { _ = resp.Body.Close() }()
	var info struct {
		Email string `json:"email"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&info)
	return info.Email
}

func ensureValidGoogleAccessToken(client *http.Client) (*googleTokenVault, error) {
	vault, err := loadGoogleTokenVault()
	if err != nil {
		return nil, err
	}
	if vault.AccessToken != "" && time.Now().Unix() < vault.ExpiresAt {
		return vault, nil
	}
	if vault.RefreshToken == "" || vault.ClientID == "" {
		return nil, fmt.Errorf("google calendar is not connected")
	}

	resp, err := client.PostForm("https://oauth2.googleapis.com/token", url.Values{
		"client_id":     {vault.ClientID},
		"refresh_token": {vault.RefreshToken},
		"grant_type":    {"refresh_token"},
	})
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to refresh google token (%d)", resp.StatusCode)
	}

	var refreshed struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&refreshed); err != nil {
		return nil, err
	}
	vault.AccessToken = refreshed.AccessToken
	vault.ExpiresAt = time.Now().Unix() + refreshed.ExpiresIn - 60
	_ = saveGoogleTokenVault(vault)
	return vault, nil
}

// FetchGoogleCalendarEvents retrieves upcoming events from the user's Google Calendar list.
func (a *App) FetchGoogleCalendarEvents(daysAhead int) ([]GoogleCalendarEventDTO, error) {
	if daysAhead <= 0 || daysAhead > 365 {
		daysAhead = 60
	}
	httpClient := &http.Client{Timeout: 15 * time.Second}
	vault, err := ensureValidGoogleAccessToken(httpClient)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	timeMin := now.AddDate(0, 0, -7).UTC().Format(time.RFC3339)
	timeMax := now.AddDate(0, 0, daysAhead).UTC().Format(time.RFC3339)

	// Fetch primary and holiday/subscribed calendars
	calListReq, _ := http.NewRequest(http.MethodGet, "https://www.googleapis.com/calendar/v3/users/me/calendarList?minAccessRole=reader", nil)
	calListReq.Header.Set("Authorization", "Bearer "+vault.AccessToken)
	calResp, err := httpClient.Do(calListReq)
	if err != nil {
		return nil, err
	}
	defer func() { _ = calResp.Body.Close() }()

	if calResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("calendarList returned %d", calResp.StatusCode)
	}

	var calList struct {
		Items []struct {
			ID       string `json:"id"`
			Summary  string `json:"summary"`
			Selected bool   `json:"selected"`
			Primary  bool   `json:"primary"`
		} `json:"items"`
	}
	if err := json.NewDecoder(calResp.Body).Decode(&calList); err != nil {
		return nil, err
	}

	var results []GoogleCalendarEventDTO
	for _, cal := range calList.Items {
		if !cal.Selected && !cal.Primary {
			continue
		}
		isHolidayCal := strings.Contains(strings.ToLower(cal.ID), "#holiday@group.v.calendar.google.com") ||
			strings.Contains(strings.ToLower(cal.Summary), "holiday")

		evURL := fmt.Sprintf(
			"https://www.googleapis.com/calendar/v3/calendars/%s/events?singleEvents=true&orderBy=startTime&maxResults=100&timeMin=%s&timeMax=%s",
			url.PathEscape(cal.ID),
			url.QueryEscape(timeMin),
			url.QueryEscape(timeMax),
		)
		evReq, _ := http.NewRequest(http.MethodGet, evURL, nil)
		evReq.Header.Set("Authorization", "Bearer "+vault.AccessToken)
		evResp, evErr := httpClient.Do(evReq)
		if evErr != nil {
			continue
		}

		var evData struct {
			Items []struct {
				ID          string `json:"id"`
				Summary     string `json:"summary"`
				Description string `json:"description"`
				Location    string `json:"location"`
				HangoutLink string `json:"hangoutLink"`
				Start       struct {
					Date     string `json:"date"`
					DateTime string `json:"dateTime"`
				} `json:"start"`
				End struct {
					Date     string `json:"date"`
					DateTime string `json:"dateTime"`
				} `json:"end"`
			} `json:"items"`
		}
		if evResp.StatusCode == http.StatusOK {
			_ = json.NewDecoder(evResp.Body).Decode(&evData)
		}
		_ = evResp.Body.Close()

		for _, item := range evData.Items {
			dateStr := item.Start.Date
			startTime := ""
			endTime := ""
			isAllDay := dateStr != ""

			if !isAllDay && item.Start.DateTime != "" {
				if parsedStart, parseErr := time.Parse(time.RFC3339, item.Start.DateTime); parseErr == nil {
					localStart := parsedStart.Local()
					dateStr = localStart.Format("2006-01-02")
					startTime = localStart.Format("15:04")
				}
				if parsedEnd, parseErr := time.Parse(time.RFC3339, item.End.DateTime); parseErr == nil {
					endTime = parsedEnd.Local().Format("15:04")
				}
			}
			if dateStr == "" {
				continue
			}

			results = append(results, GoogleCalendarEventDTO{
				ID:          item.ID,
				Title:       strings.TrimSpace(item.Summary),
				Description: strings.TrimSpace(item.Description),
				Date:        dateStr,
				StartTime:   startTime,
				EndTime:     endTime,
				IsAllDay:    isAllDay,
				Location:    strings.TrimSpace(item.Location),
				HangoutLink: strings.TrimSpace(item.HangoutLink),
				IsHoliday:   isHolidayCal,
				CalendarID:  cal.ID,
			})
		}
	}

	vault.LastSyncedAt = time.Now().UnixMilli()
	_ = saveGoogleTokenVault(vault)
	return results, nil
}
