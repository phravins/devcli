package history

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type Entry struct {
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	CreatedAt time.Time `json:"created_at"`
}

func getHistoryPath() string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".devcli")
	os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "history.json")
}

func Load() ([]Entry, error) {
	path := getHistoryPath()
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return []Entry{}, nil
	}
	if err != nil {
		return nil, err
	}

	var entries []Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return []Entry{}, nil
	}
	return entries, nil
}

func Save(entries []Entry) error {
	path := getHistoryPath()
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func Add(name, path string) error {
	entries, _ := Load()
	// Prepend new entry
	newEntry := Entry{
		Name:      name,
		Path:      path,
		CreatedAt: time.Now(),
	}
	entries = append([]Entry{newEntry}, entries...)
	return Save(entries)
}

func GetOldEntries(days int) []Entry {
	entries, _ := Load()
	cutoff := time.Now().AddDate(0, 0, -days)
	var old []Entry
	for _, e := range entries {
		if e.CreatedAt.Before(cutoff) {
			old = append(old, e)
		}
	}
	return old
}

func DeleteOld(days int) error {
	entries, _ := Load()
	cutoff := time.Now().AddDate(0, 0, -days)
	var kept []Entry
	for _, e := range entries {
		if !e.CreatedAt.Before(cutoff) {
			kept = append(kept, e)
		}
	}
	return Save(kept)
}

func DeleteOne(index int) error {
	entries, _ := Load()
	if index < 0 || index >= len(entries) {
		return nil
	}
	// Remove entry at index
	entries = append(entries[:index], entries[index+1:]...)
	return Save(entries)
}
