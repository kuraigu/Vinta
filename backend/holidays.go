// Package backend provides core application services and state management for Vinta.
package backend

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// PublicHoliday represents a single public holiday entry from Nager.Date.
type PublicHoliday struct {
	Date        string   `json:"date"`
	LocalName   string   `json:"localName"`
	Name        string   `json:"name"`
	CountryCode string   `json:"countryCode"`
	Fixed       bool     `json:"fixed"`
	Global      bool     `json:"global"`
	Types       []string `json:"types"`
}

// HolidayCountry describes a supported country in the Nager.Date public holidays API.
type HolidayCountry struct {
	CountryCode string `json:"countryCode"`
	Name        string `json:"name"`
}

// GetPublicHolidays fetches public holidays for the specified year and ISO country code,
// using an annual disk cache under %APPDATA%/FreeMatrix Studios/Vinta/cache.
func (a *App) GetPublicHolidays(year int, countryCode string) ([]PublicHoliday, error) {
	code := strings.ToUpper(strings.TrimSpace(countryCode))
	if code == "" {
		code = "US"
	}
	if year < 1970 || year > 2100 {
		year = time.Now().Year()
	}

	cacheDir := filepath.Join(GetAppDataDir(), "cache")
	cachePath := filepath.Join(cacheDir, fmt.Sprintf("holidays_%s_%d.json", code, year))

	if info, err := os.Stat(cachePath); err == nil {
		// Re-use past/future year caches indefinitely, and current-year cache for 30 days.
		if year != time.Now().Year() || time.Since(info.ModTime()) < 30*24*time.Hour {
			if raw, readErr := os.ReadFile(cachePath); readErr == nil && len(raw) > 0 {
				var cached []PublicHoliday
				if json.Unmarshal(raw, &cached) == nil {
					return cached, nil
				}
			}
		}
	}

	client := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("https://date.nager.at/api/v3/PublicHolidays/%d/%s", year, code)
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch holidays: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("holidays API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("failed to read holidays response: %w", err)
	}

	var holidays []PublicHoliday
	if err := json.Unmarshal(body, &holidays); err != nil {
		return nil, fmt.Errorf("failed to decode holidays: %w", err)
	}

	if err := os.MkdirAll(cacheDir, 0755); err == nil {
		_ = os.WriteFile(cachePath, body, 0644)
	}

	return holidays, nil
}

// GetAvailableHolidayCountries returns supported countries for Nager.Date public holidays,
// cached for 30 days under %APPDATA%/FreeMatrix Studios/Vinta/cache.
func (a *App) GetAvailableHolidayCountries() ([]HolidayCountry, error) {
	cacheDir := filepath.Join(GetAppDataDir(), "cache")
	cachePath := filepath.Join(cacheDir, "holiday_countries.json")

	if info, err := os.Stat(cachePath); err == nil && time.Since(info.ModTime()) < 30*24*time.Hour {
		if raw, readErr := os.ReadFile(cachePath); readErr == nil && len(raw) > 0 {
			var cached []HolidayCountry
			if json.Unmarshal(raw, &cached) == nil {
				return cached, nil
			}
		}
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://date.nager.at/api/v3/AvailableCountries")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch holiday countries: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("holiday countries API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if err != nil {
		return nil, fmt.Errorf("failed to read countries response: %w", err)
	}

	var countries []HolidayCountry
	if err := json.Unmarshal(body, &countries); err != nil {
		return nil, fmt.Errorf("failed to decode countries: %w", err)
	}

	if err := os.MkdirAll(cacheDir, 0755); err == nil {
		_ = os.WriteFile(cachePath, body, 0644)
	}

	return countries, nil
}
