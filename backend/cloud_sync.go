// Package backend provides core application services and state management for Vinta.
package backend

import (
	"bytes"
	"context"
	"crypto/rand"
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

// GoogleAccountStatus reports the active Google Sign-In and Firebase Cloud Sync state.
type GoogleAccountStatus struct {
	Connected         bool   `json:"connected"`
	Email             string `json:"email"`
	Name              string `json:"name"`
	Picture           string `json:"picture"`
	HasGoogleClientID bool   `json:"hasGoogleClientId"`
	FirebaseEnabled   bool   `json:"firebaseEnabled"`
	CloudSyncReady    bool   `json:"cloudSyncReady"`
	LastCloudSyncAt   int64  `json:"lastCloudSyncAt"`
}

// CloudSyncResult reports the outcome of a Firestore push or pull operation.
type CloudSyncResult struct {
	SyncedAt int64    `json:"syncedAt"`
	Files    []string `json:"files"`
	Message  string   `json:"message"`
}

var cloudSyncFileNames = []string{
	"settings.json",
	"calendar.json",
	"flashcards.json",
	"habits.json",
}

// GetGoogleAccountStatus returns the current Google account and Firebase Cloud Sync readiness.
func (a *App) GetGoogleAccountStatus() GoogleAccountStatus {
	vault, _ := readCredentialVault()
	hasGoogle := GetEnvConfig("VINTA_GOOGLE_CLIENT_ID") != ""
	hasFirebase := GetEnvConfig("VINTA_FIREBASE_API_KEY") != "" && GetEnvConfig("VINTA_FIREBASE_PROJECT_ID") != ""
	connected := strings.TrimSpace(vault.Google.AccessToken) != "" || strings.TrimSpace(vault.Google.RefreshToken) != ""
	cloudReady := connected && hasFirebase && strings.TrimSpace(vault.Google.FirebaseUID) != "" &&
		(strings.TrimSpace(vault.Google.FirebaseIDToken) != "" || strings.TrimSpace(vault.Google.FirebaseRefreshToken) != "")

	return GoogleAccountStatus{
		Connected:         connected,
		Email:             vault.Google.Email,
		Name:              vault.Google.Name,
		Picture:           vault.Google.Picture,
		HasGoogleClientID: hasGoogle,
		FirebaseEnabled:   hasFirebase,
		CloudSyncReady:    cloudReady,
		LastCloudSyncAt:   vault.Google.LastCloudSyncAt,
	}
}

// SignInWithGoogle runs one RFC 8252 loopback PKCE flow that powers both Firebase Cloud Sync and Google Calendar.
func (a *App) SignInWithGoogle() (GoogleAccountStatus, error) {
	clientID := GetEnvConfig("VINTA_GOOGLE_CLIENT_ID")
	if clientID == "" {
		return a.GetGoogleAccountStatus(), fmt.Errorf("google sign-in is not configured in this build")
	}

	verifier, challenge, err := generatePKCEPair()
	if err != nil {
		return a.GetGoogleAccountStatus(), err
	}

	stateBuf := make([]byte, 16)
	if _, err := rand.Read(stateBuf); err != nil {
		return a.GetGoogleAccountStatus(), err
	}
	oauthState := base64.RawURLEncoding.EncodeToString(stateBuf)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return a.GetGoogleAccountStatus(), fmt.Errorf("failed to start OAuth callback listener: %w", err)
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
			http.Error(w, "Sign-in cancelled", http.StatusBadRequest)
			errCh <- fmt.Errorf("google sign-in cancelled: %s", oauthErr)
			return
		}
		code := q.Get("code")
		if code == "" {
			http.Error(w, "Missing authorization code", http.StatusBadRequest)
			errCh <- fmt.Errorf("missing authorization code")
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<!doctype html><html><body style="font-family:system-ui;background:#0f172a;color:#f8fafc;display:flex;align-items:center;justify-content:center;height:100vh;margin:0"><div style="text-align:center"><h2>Signed in to Vinta</h2><p>You can close this browser tab and return to Vinta.</p></div></body></html>`))
		codeCh <- code
	})

	srv := &http.Server{Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	go func() {
		_ = srv.Serve(ln)
	}()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}()

	authParams := url.Values{}
	authParams.Set("client_id", clientID)
	authParams.Set("redirect_uri", redirectURI)
	authParams.Set("response_type", "code")
	authParams.Set("scope", "openid email profile https://www.googleapis.com/auth/calendar.readonly")
	authParams.Set("access_type", "offline")
	authParams.Set("prompt", "consent")
	authParams.Set("state", oauthState)
	authParams.Set("code_challenge", challenge)
	authParams.Set("code_challenge_method", "S256")

	if err := a.OpenURL("https://accounts.google.com/o/oauth2/v2/auth?" + authParams.Encode()); err != nil {
		return a.GetGoogleAccountStatus(), fmt.Errorf("failed to open browser: %w", err)
	}

	var authCode string
	select {
	case authCode = <-codeCh:
	case cbErr := <-errCh:
		return a.GetGoogleAccountStatus(), cbErr
	case <-time.After(3 * time.Minute):
		return a.GetGoogleAccountStatus(), fmt.Errorf("google sign-in timed out")
	}

	form := url.Values{}
	form.Set("client_id", clientID)
	form.Set("code", authCode)
	form.Set("code_verifier", verifier)
	form.Set("grant_type", "authorization_code")
	form.Set("redirect_uri", redirectURI)

	tokenResp, err := sharedAPIClient.PostForm("https://oauth2.googleapis.com/token", form)
	if err != nil {
		return a.GetGoogleAccountStatus(), fmt.Errorf("google token exchange failed: %w", err)
	}
	defer func() { _ = tokenResp.Body.Close() }()

	tokenBody, _ := io.ReadAll(io.LimitReader(tokenResp.Body, 1024*1024))
	if tokenResp.StatusCode != http.StatusOK {
		return a.GetGoogleAccountStatus(), fmt.Errorf("google token error %d", tokenResp.StatusCode)
	}

	var tok struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		IDToken      string `json:"id_token"`
		ExpiresIn    int64  `json:"expires_in"`
	}
	if err := json.Unmarshal(tokenBody, &tok); err != nil {
		return a.GetGoogleAccountStatus(), err
	}

	profileEmail, profileName, profilePicture := fetchGoogleUserProfile(tok.AccessToken)
	expiresAt := time.Now().Unix() + tok.ExpiresIn - 60

	var fbIDToken, fbRefreshToken, fbUID string
	var fbExpiresAt int64
	if tok.IDToken != "" && GetEnvConfig("VINTA_FIREBASE_API_KEY") != "" {
		fbIDToken, fbRefreshToken, fbUID, fbExpiresAt, _ = exchangeGoogleForFirebase(tok.IDToken)
	}

	if err := updateCredentialVault(func(v *credentialVault) {
		v.Google.Email = profileEmail
		v.Google.Name = profileName
		v.Google.Picture = profilePicture
		v.Google.IDToken = tok.IDToken
		v.Google.AccessToken = tok.AccessToken
		if tok.RefreshToken != "" {
			v.Google.RefreshToken = tok.RefreshToken
		}
		v.Google.ExpiresAt = expiresAt
		if fbUID != "" {
			v.Google.FirebaseIDToken = fbIDToken
			v.Google.FirebaseRefreshToken = fbRefreshToken
			v.Google.FirebaseUID = fbUID
			v.Google.FirebaseExpiresAt = fbExpiresAt
		}
	}); err != nil {
		return a.GetGoogleAccountStatus(), err
	}

	// Mirror Google OAuth tokens into Google Calendar vault so Calendar sync unlocks automatically.
	_ = saveGoogleTokenVault(&googleTokenVault{
		ClientID:     clientID,
		AccessToken:  tok.AccessToken,
		RefreshToken: tok.RefreshToken,
		ExpiresAt:    expiresAt,
		Email:        profileEmail,
	})

	return a.GetGoogleAccountStatus(), nil
}

// SignOutGoogle clears Google and Firebase credentials from DPAPI storage and disconnects Google Calendar.
func (a *App) SignOutGoogle() error {
	_ = a.DisconnectGoogleCalendar()
	return updateCredentialVault(func(v *credentialVault) {
		v.Google = googleCredentialRecord{}
	})
}

func fetchGoogleUserProfile(accessToken string) (email, name, picture string) {
	req, err := http.NewRequest(http.MethodGet, "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return "", "", ""
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := sharedAPIClient.Do(req)
	if err != nil {
		return "", "", ""
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", "", ""
	}
	var info struct {
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err == nil {
		return info.Email, info.Name, info.Picture
	}
	return "", "", ""
}

func exchangeGoogleForFirebase(googleIDToken string) (idToken, refreshToken, uid string, expiresAt int64, err error) {
	apiKey := GetEnvConfig("VINTA_FIREBASE_API_KEY")
	if apiKey == "" {
		return "", "", "", 0, fmt.Errorf("firebase api key not set")
	}
	payload, _ := json.Marshal(map[string]any{
		"postBody":            "id_token=" + url.QueryEscape(googleIDToken) + "&providerId=google.com",
		"requestUri":          "http://127.0.0.1",
		"returnIdpCredential": true,
		"returnSecureToken":   true,
	})
	endpoint := "https://identitytoolkit.googleapis.com/v1/accounts:signInWithIdp?key=" + url.QueryEscape(apiKey)
	resp, err := sharedAPIClient.Post(endpoint, "application/json", bytes.NewReader(payload))
	if err != nil {
		return "", "", "", 0, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if resp.StatusCode != http.StatusOK {
		return "", "", "", 0, fmt.Errorf("firebase signInWithIdp status %d", resp.StatusCode)
	}
	var res struct {
		IDToken      string `json:"idToken"`
		RefreshToken string `json:"refreshToken"`
		LocalID      string `json:"localId"`
		ExpiresIn    string `json:"expiresIn"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return "", "", "", 0, err
	}
	return res.IDToken, res.RefreshToken, res.LocalID, time.Now().Unix() + 3500, nil
}

func ensureFreshFirebaseToken() (idToken string, uid string, err error) {
	vault, err := readCredentialVault()
	if err != nil {
		return "", "", err
	}
	if strings.TrimSpace(vault.Google.FirebaseUID) == "" {
		return "", "", fmt.Errorf("not signed in to Firebase")
	}
	if strings.TrimSpace(vault.Google.FirebaseIDToken) != "" && time.Now().Unix() < vault.Google.FirebaseExpiresAt {
		return vault.Google.FirebaseIDToken, vault.Google.FirebaseUID, nil
	}
	if strings.TrimSpace(vault.Google.FirebaseRefreshToken) == "" {
		return "", "", fmt.Errorf("missing firebase refresh token; please sign in again")
	}
	apiKey := GetEnvConfig("VINTA_FIREBASE_API_KEY")
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", vault.Google.FirebaseRefreshToken)
	resp, err := sharedAPIClient.PostForm("https://securetoken.googleapis.com/v1/token?key="+url.QueryEscape(apiKey), form)
	if err != nil {
		return "", "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("firebase token refresh failed (%d)", resp.StatusCode)
	}
	var refreshed struct {
		IDToken      string `json:"id_token"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&refreshed); err != nil {
		return "", "", err
	}
	_ = updateCredentialVault(func(v *credentialVault) {
		v.Google.FirebaseIDToken = refreshed.IDToken
		if refreshed.RefreshToken != "" {
			v.Google.FirebaseRefreshToken = refreshed.RefreshToken
		}
		v.Google.FirebaseExpiresAt = time.Now().Unix() + 3500
	})
	return refreshed.IDToken, vault.Google.FirebaseUID, nil
}

// PushCloudSync uploads local settings and companion data to the user's Firestore backup document.
func (a *App) PushCloudSync() (CloudSyncResult, error) {
	projectID := GetEnvConfig("VINTA_FIREBASE_PROJECT_ID")
	if projectID == "" {
		return CloudSyncResult{}, fmt.Errorf("VINTA_FIREBASE_PROJECT_ID is not configured")
	}
	idToken, uid, err := ensureFreshFirebaseToken()
	if err != nil {
		return CloudSyncResult{}, err
	}

	dir := GetAppDataDir()
	fields := make(map[string]any)
	syncedFiles := make([]string, 0, len(cloudSyncFileNames))
	for _, name := range cloudSyncFileNames {
		data, readErr := os.ReadFile(filepath.Join(dir, name))
		if readErr != nil || len(data) == 0 {
			continue
		}
		fields[strings.TrimSuffix(name, ".json")] = map[string]string{"stringValue": string(data)}
		syncedFiles = append(syncedFiles, name)
	}
	now := time.Now().Unix()
	fields["updatedAt"] = map[string]string{"integerValue": fmt.Sprintf("%d", now)}

	payload, _ := json.Marshal(map[string]any{"fields": fields})
	docURL := fmt.Sprintf("https://firestore.googleapis.com/v1/projects/%s/databases/(default)/documents/users/%s/vinta/backup", url.PathEscape(projectID), url.PathEscape(uid))

	req, err := http.NewRequest(http.MethodPatch, docURL, bytes.NewReader(payload))
	if err != nil {
		return CloudSyncResult{}, err
	}
	req.Header.Set("Authorization", "Bearer "+idToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := sharedAPIClient.Do(req)
	if err != nil {
		return CloudSyncResult{}, fmt.Errorf("cloud sync upload failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return CloudSyncResult{}, fmt.Errorf("firestore returned status %d", resp.StatusCode)
	}

	_ = updateCredentialVault(func(v *credentialVault) {
		v.Google.LastCloudSyncAt = now
	})
	return CloudSyncResult{
		SyncedAt: now,
		Files:    syncedFiles,
		Message:  "Cloud backup updated",
	}, nil
}

// PullCloudSync restores settings and companion data from Firestore after taking a local safety backup.
func (a *App) PullCloudSync() (CloudSyncResult, error) {
	projectID := GetEnvConfig("VINTA_FIREBASE_PROJECT_ID")
	if projectID == "" {
		return CloudSyncResult{}, fmt.Errorf("VINTA_FIREBASE_PROJECT_ID is not configured")
	}
	idToken, uid, err := ensureFreshFirebaseToken()
	if err != nil {
		return CloudSyncResult{}, err
	}

	docURL := fmt.Sprintf("https://firestore.googleapis.com/v1/projects/%s/databases/(default)/documents/users/%s/vinta/backup", url.PathEscape(projectID), url.PathEscape(uid))
	req, err := http.NewRequest(http.MethodGet, docURL, nil)
	if err != nil {
		return CloudSyncResult{}, err
	}
	req.Header.Set("Authorization", "Bearer "+idToken)

	resp, err := sharedAPIClient.Do(req)
	if err != nil {
		return CloudSyncResult{}, fmt.Errorf("cloud sync download failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusNotFound {
		return CloudSyncResult{}, fmt.Errorf("no cloud backup found for this account yet")
	}
	if resp.StatusCode != http.StatusOK {
		return CloudSyncResult{}, fmt.Errorf("firestore returned status %d", resp.StatusCode)
	}

	var doc struct {
		Fields map[string]struct {
			StringValue string `json:"stringValue"`
		} `json:"fields"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return CloudSyncResult{}, err
	}

	// Snapshot local settings before overwriting from cloud.
	_, _ = a.CreateSettingsBackup()

	dir := GetAppDataDir()
	_ = os.MkdirAll(dir, 0755)
	restored := make([]string, 0, len(cloudSyncFileNames))
	for _, name := range cloudSyncFileNames {
		key := strings.TrimSuffix(name, ".json")
		field, ok := doc.Fields[key]
		if !ok || strings.TrimSpace(field.StringValue) == "" {
			continue
		}
		if err := os.WriteFile(filepath.Join(dir, name), []byte(field.StringValue), 0600); err == nil {
			restored = append(restored, name)
		}
	}

	now := time.Now().Unix()
	_ = updateCredentialVault(func(v *credentialVault) {
		v.Google.LastCloudSyncAt = now
	})
	return CloudSyncResult{
		SyncedAt: now,
		Files:    restored,
		Message:  "Cloud backup restored",
	}, nil
}
