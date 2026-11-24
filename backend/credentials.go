// Package backend provides core application services and state management for Vinta.
package backend

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// AIKeysStatus exposes presence flags for DPAPI-stored AI provider keys without leaking secrets.
type AIKeysStatus struct {
	HasClaude bool `json:"hasClaude"`
	HasOpenAI bool `json:"hasOpenAI"`
}

type googleCredentialRecord struct {
	Email                string `json:"email"`
	Name                 string `json:"name"`
	Picture              string `json:"picture"`
	IDToken              string `json:"idToken"`
	AccessToken          string `json:"accessToken"`
	RefreshToken         string `json:"refreshToken"`
	ExpiresAt            int64  `json:"expiresAt"`
	FirebaseIDToken      string `json:"firebaseIdToken"`
	FirebaseRefreshToken string `json:"firebaseRefreshToken"`
	FirebaseUID          string `json:"firebaseUid"`
	FirebaseExpiresAt    int64  `json:"firebaseExpiresAt"`
	LastCloudSyncAt      int64  `json:"lastCloudSyncAt"`
}

type githubCredentialRecord struct {
	AccessToken string `json:"accessToken"`
	Username    string `json:"username"`
	AvatarURL   string `json:"avatarUrl"`
	ConnectedAt int64  `json:"connectedAt"`
}

type aiKeysRecord struct {
	ClaudeKey string `json:"claudeKey"`
	OpenAIKey string `json:"openAiKey"`
}

type credentialVault struct {
	Google googleCredentialRecord `json:"google"`
	GitHub githubCredentialRecord `json:"github"`
	AIKeys aiKeysRecord           `json:"aiKeys"`
}

var (
	dpapiCredVaultHeader = []byte("VNCR")
	credVaultMu          sync.RWMutex
)

func getCredentialVaultPath() string {
	return filepath.Join(GetAppDataDir(), "credentials.bin")
}

func loadCredentialVaultLocked() (*credentialVault, error) {
	raw, err := os.ReadFile(getCredentialVaultPath())
	if err != nil {
		if os.IsNotExist(err) {
			return &credentialVault{}, nil
		}
		return &credentialVault{}, err
	}
	if len(raw) == 0 {
		return &credentialVault{}, nil
	}
	if len(raw) >= len(dpapiCredVaultHeader) && string(raw[:len(dpapiCredVaultHeader)]) == string(dpapiCredVaultHeader) {
		decrypted, decErr := decryptDPAPI(raw[len(dpapiCredVaultHeader):])
		if decErr != nil {
			return &credentialVault{}, decErr
		}
		raw = decrypted
	}
	var vault credentialVault
	if err := json.Unmarshal(raw, &vault); err != nil {
		return &credentialVault{}, err
	}
	return &vault, nil
}

func saveCredentialVaultLocked(vault *credentialVault) error {
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
	out := append(append([]byte{}, dpapiCredVaultHeader...), encrypted...)
	return os.WriteFile(getCredentialVaultPath(), out, 0600)
}

func readCredentialVault() (*credentialVault, error) {
	credVaultMu.RLock()
	defer credVaultMu.RUnlock()
	return loadCredentialVaultLocked()
}

func updateCredentialVault(mutate func(*credentialVault)) error {
	credVaultMu.Lock()
	defer credVaultMu.Unlock()
	vault, err := loadCredentialVaultLocked()
	if err != nil {
		vault = &credentialVault{}
	}
	mutate(vault)
	return saveCredentialVaultLocked(vault)
}

// GetAIKeysStatus reports which AI provider API keys are currently stored in the DPAPI vault.
func (a *App) GetAIKeysStatus() AIKeysStatus {
	vault, err := readCredentialVault()
	if err != nil {
		return AIKeysStatus{}
	}
	return AIKeysStatus{
		HasClaude: strings.TrimSpace(vault.AIKeys.ClaudeKey) != "",
		HasOpenAI: strings.TrimSpace(vault.AIKeys.OpenAIKey) != "",
	}
}

// SaveAIProviderKey stores an AI provider API key encrypted via Windows DPAPI.
func (a *App) SaveAIProviderKey(provider, apiKey string) error {
	trimmed := strings.TrimSpace(apiKey)
	if trimmed == "" {
		return fmt.Errorf("api key cannot be empty")
	}
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "claude", "anthropic":
		return updateCredentialVault(func(v *credentialVault) { v.AIKeys.ClaudeKey = trimmed })
	case "openai":
		return updateCredentialVault(func(v *credentialVault) { v.AIKeys.OpenAIKey = trimmed })
	default:
		return fmt.Errorf("unsupported AI provider: %s", provider)
	}
}

// DeleteAIProviderKey removes a stored AI provider API key from the DPAPI vault.
func (a *App) DeleteAIProviderKey(provider string) error {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "claude", "anthropic":
		return updateCredentialVault(func(v *credentialVault) { v.AIKeys.ClaudeKey = "" })
	case "openai":
		return updateCredentialVault(func(v *credentialVault) { v.AIKeys.OpenAIKey = "" })
	default:
		return fmt.Errorf("unsupported AI provider: %s", provider)
	}
}
