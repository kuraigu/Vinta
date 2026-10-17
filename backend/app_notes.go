// Package backend provides core application services and state management for Vinta.
package backend

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// NoteFileInfo captures summary metadata for a saved sticky note file on disk.
type NoteFileInfo struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	FileName string `json:"fileName"`
	ModTime  int64  `json:"modTime"`
	Size     int64  `json:"size"`
}

type noteTitleParser struct {
	Title string `json:"title"`
}

// getNotesDir resolves and initializes the notes directory under %APPDATA%/FreeMatrix Studios/Vinta/notes.
func getNotesDir() (string, error) {
	notesDir := filepath.Join(GetAppDataDir(), "notes")
	if err := os.MkdirAll(notesDir, 0755); err != nil {
		return "", err
	}
	return notesDir, nil
}

// migrateLegacyNotes checks for legacy .md notes and converts them into .vnote JSON format.
func migrateLegacyNotes(dir string, entries []os.DirEntry) {
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".md") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		oldPath := filepath.Join(dir, entry.Name())
		mdContent, err := os.ReadFile(oldPath)
		if err != nil {
			continue
		}

		baseName := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		newFileName := baseName + ".vnote"
		newPath := filepath.Join(dir, newFileName)

		if _, statErr := os.Stat(newPath); os.IsNotExist(statErr) {
			payload := map[string]interface{}{
				"id":        baseName,
				"title":     baseName,
				"color":     "yellow",
				"content":   string(mdContent),
				"pinned":    false,
				"modTime":   info.ModTime().UnixMilli(),
				"createdAt": info.ModTime().UnixMilli(),
			}
			if data, err := json.MarshalIndent(payload, "", "  "); err == nil {
				_ = os.WriteFile(newPath, data, 0644)
			}
		}

		_ = os.Remove(oldPath)
	}
}

// ListNotes scans the notes storage directory, runs migration if necessary, and returns all note metadata.
func (a *App) ListNotes() ([]NoteFileInfo, error) {
	dir, err := getNotesDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	migrateLegacyNotes(dir, entries)

	entries, err = os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var notes []NoteFileInfo
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".vnote") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		fileName := entry.Name()
		id := strings.TrimSuffix(fileName, filepath.Ext(fileName))
		title := id

		filePath := filepath.Join(dir, fileName)
		if fileBytes, readErr := os.ReadFile(filePath); readErr == nil {
			var parsed noteTitleParser
			if parseErr := json.Unmarshal(fileBytes, &parsed); parseErr == nil && strings.TrimSpace(parsed.Title) != "" {
				title = strings.TrimSpace(parsed.Title)
			}
		}

		notes = append(notes, NoteFileInfo{
			ID:       id,
			Title:    title,
			FileName: fileName,
			ModTime:  info.ModTime().UnixMilli(),
			Size:     info.Size(),
		})
	}
	return notes, nil
}

// ReadNote reads and returns the JSON content of the requested note file.
func (a *App) ReadNote(fileName string) (string, error) {
	dir, err := getNotesDir()
	if err != nil {
		return "", err
	}
	cleanName := filepath.Base(fileName)
	if !strings.HasSuffix(strings.ToLower(cleanName), ".vnote") {
		cleanName = strings.TrimSuffix(cleanName, filepath.Ext(cleanName)) + ".vnote"
	}
	filePath := filepath.Join(dir, cleanName)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// SaveNote writes serialized note JSON content to the specified note file.
func (a *App) SaveNote(fileName string, content string) error {
	dir, err := getNotesDir()
	if err != nil {
		return err
	}
	cleanName := filepath.Base(fileName)
	if !strings.HasSuffix(strings.ToLower(cleanName), ".vnote") {
		cleanName = strings.TrimSuffix(cleanName, filepath.Ext(cleanName)) + ".vnote"
	}
	filePath := filepath.Join(dir, cleanName)
	return os.WriteFile(filePath, []byte(content), 0644)
}

// RenameNote renames an existing note file using the sanitized title as the new filename.
func (a *App) RenameNote(oldFileName string, newTitle string) (string, error) {
	dir, err := getNotesDir()
	if err != nil {
		return "", err
	}
	sanitized := strings.NewReplacer(
		"/", "-", "\\", "-", ":", "-", "*", "-",
		"?", "-", "\"", "-", "<", "-", ">", "-", "|", "-",
	).Replace(strings.TrimSpace(newTitle))
	if sanitized == "" {
		sanitized = "untitled"
	}
	newFileName := sanitized + ".vnote"

	oldClean := filepath.Base(oldFileName)
	if !strings.HasSuffix(strings.ToLower(oldClean), ".vnote") {
		oldClean = strings.TrimSuffix(oldClean, filepath.Ext(oldClean)) + ".vnote"
	}

	oldPath := filepath.Join(dir, oldClean)
	newPath := filepath.Join(dir, newFileName)

	if !strings.EqualFold(oldPath, newPath) {
		if _, statErr := os.Stat(newPath); statErr == nil {
			base := strings.TrimSuffix(newFileName, ".vnote")
			for n := 2; ; n++ {
				candidate := fmt.Sprintf("%s (%d).vnote", base, n)
				if _, e := os.Stat(filepath.Join(dir, candidate)); os.IsNotExist(e) {
					newFileName = candidate
					newPath = filepath.Join(dir, newFileName)
					break
				}
			}
		}
		if err := os.Rename(oldPath, newPath); err != nil {
			return "", err
		}
	}
	return newFileName, nil
}

// DeleteNote removes a note file from the storage directory.
func (a *App) DeleteNote(fileName string) error {
	dir, err := getNotesDir()
	if err != nil {
		return err
	}
	cleanName := filepath.Base(fileName)
	if !strings.HasSuffix(strings.ToLower(cleanName), ".vnote") {
		cleanName = strings.TrimSuffix(cleanName, filepath.Ext(cleanName)) + ".vnote"
	}
	filePath := filepath.Join(dir, cleanName)
	return os.Remove(filePath)
}

// OpenNotesFolder launches Windows Explorer targeting the notes storage directory.
func (a *App) OpenNotesFolder() error {
	dir, err := getNotesDir()
	if err != nil {
		return err
	}
	return exec.Command("explorer.exe", dir).Start()
}
