package snippets

import (
	"os"
	"testing"
	"time"
)

func setupTestStorage(t *testing.T) (*Storage, string) {
	t.Helper()
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("USERPROFILE", tmpDir)

	storage, err := NewStorage()
	if err != nil {
		t.Fatalf("NewStorage failed: %v", err)
	}
	return storage, tmpDir
}

func TestNewStorage_Error(t *testing.T) {
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "")

	_, err := NewStorage()
	if err == nil {
		t.Fatal("expected error when HOME is not set, got nil")
	}
}

func TestLoadAll_FileNotExist(t *testing.T) {
	storage, _ := setupTestStorage(t)

	snippets, err := storage.LoadAll()
	if err != nil {
		t.Fatalf("expected no error when file does not exist, got: %v", err)
	}
	if len(snippets) != 0 {
		t.Errorf("expected 0 snippets, got %d", len(snippets))
	}
}

func TestLoadAll_Success(t *testing.T) {
	storage, _ := setupTestStorage(t)

	expected := []Snippet{
		{
			ID:          "test-1",
			Title:       "Test Title",
			Description: "Test Description",
			Language:    "go",
			Category:    "testing",
			Code:        "fmt.Println(\"hello\")",
			Tags:        []string{"go", "test"},
			CreatedAt:   time.Now().Truncate(time.Second),
			UpdatedAt:   time.Now().Truncate(time.Second),
		},
	}

	if err := storage.SaveAll(expected); err != nil {
		t.Fatalf("SaveAll failed: %v", err)
	}

	loaded, err := storage.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}

	if len(loaded) != 1 {
		t.Fatalf("expected 1 snippet, got %d", len(loaded))
	}

	if loaded[0].ID != expected[0].ID || loaded[0].Title != expected[0].Title || loaded[0].Code != expected[0].Code {
		t.Errorf("loaded snippet %+v does not match expected %+v", loaded[0], expected[0])
	}
}

func TestLoadAll_InvalidJSON(t *testing.T) {
	storage, _ := setupTestStorage(t)

	// Write invalid JSON to filePath
	if err := os.WriteFile(storage.filePath, []byte("{invalid json"), 0644); err != nil {
		t.Fatalf("failed to write invalid json: %v", err)
	}

	_, err := storage.LoadAll()
	if err == nil {
		t.Fatal("expected error when loading invalid JSON, got nil")
	}
}

func TestLoadAll_ReadError(t *testing.T) {
	storage, _ := setupTestStorage(t)

	// Create a directory at s.filePath so os.ReadFile fails
	if err := os.MkdirAll(storage.filePath, 0755); err != nil {
		t.Fatalf("failed to create directory at filePath: %v", err)
	}

	_, err := storage.LoadAll()
	if err == nil {
		t.Fatal("expected error when filePath is a directory, got nil")
	}
}

func TestStorage_CRUD(t *testing.T) {
	storage, _ := setupTestStorage(t)

	// Add snippet
	snip := Snippet{
		Title:       "Bubble Tea App",
		Description: "TUI app template",
		Language:    "go",
		Category:    "tui",
		Code:        "tea.NewProgram(...)",
		Tags:        []string{"charm", "tui"},
	}

	err := storage.Add(snip)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	all, err := storage.LoadAll()
	if err != nil || len(all) != 1 {
		t.Fatalf("expected 1 snippet after Add, got %d (err: %v)", len(all), err)
	}

	addedID := all[0].ID
	if addedID == "" {
		t.Error("expected non-empty snippet ID")
	}

	// Get snippet
	retrieved, err := storage.Get(addedID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if retrieved.Title != "Bubble Tea App" {
		t.Errorf("expected title 'Bubble Tea App', got '%s'", retrieved.Title)
	}

	// Get non-existent
	_, err = storage.Get("non-existent-id")
	if err == nil {
		t.Error("expected error getting non-existent snippet, got nil")
	}

	// Update snippet
	updatedSnip := *retrieved
	updatedSnip.Title = "Updated Bubble Tea App"
	err = storage.Update(updatedSnip)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	retrievedUpdated, err := storage.Get(addedID)
	if err != nil {
		t.Fatalf("Get after Update failed: %v", err)
	}
	if retrievedUpdated.Title != "Updated Bubble Tea App" {
		t.Errorf("expected updated title 'Updated Bubble Tea App', got '%s'", retrievedUpdated.Title)
	}

	// Update non-existent
	badUpdate := Snippet{ID: "unknown-id", Title: "Foo"}
	if err := storage.Update(badUpdate); err == nil {
		t.Error("expected error updating non-existent snippet, got nil")
	}

	// Delete snippet
	err = storage.Delete(addedID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = storage.Get(addedID)
	if err == nil {
		t.Error("expected error getting deleted snippet, got nil")
	}

	// Delete non-existent
	if err := storage.Delete("unknown-id"); err == nil {
		t.Error("expected error deleting non-existent snippet, got nil")
	}
}

func TestStorage_SearchAndFilter(t *testing.T) {
	storage, _ := setupTestStorage(t)

	s1 := Snippet{Title: "Go Web Server", Description: "HTTP web server", Language: "go", Category: "api", Code: "http.ListenAndServe", Tags: []string{"web", "http"}}
	s2 := Snippet{Title: "Python Flask", Description: "Micro framework", Language: "python", Category: "api", Code: "app.run()", Tags: []string{"web", "flask"}}
	s3 := Snippet{Title: "React Hook", Description: "Custom state hook", Language: "javascript", Category: "ui", Code: "useState()", Tags: []string{"react", "hook"}}

	_ = storage.Add(s1)
	_ = storage.Add(s2)
	_ = storage.Add(s3)

	// Search
	results, err := storage.Search("flask")
	if err != nil || len(results) != 1 {
		t.Fatalf("expected 1 search result for 'flask', got %d (err: %v)", len(results), err)
	}
	if results[0].Title != "Python Flask" {
		t.Errorf("unexpected search result: %v", results[0].Title)
	}

	// Search by tag
	results, err = storage.Search("react")
	if err != nil || len(results) != 1 {
		t.Fatalf("expected 1 search result for tag 'react', got %d", len(results))
	}

	// FilterByCategory
	apiSnippets, err := storage.FilterByCategory("API") // case-insensitive check
	if err != nil || len(apiSnippets) != 2 {
		t.Fatalf("expected 2 API category snippets, got %d (err: %v)", len(apiSnippets), err)
	}

	// FilterByLanguage
	goSnippets, err := storage.FilterByLanguage("GO") // case-insensitive check
	if err != nil || len(goSnippets) != 1 {
		t.Fatalf("expected 1 Go snippet, got %d (err: %v)", len(goSnippets), err)
	}
	if goSnippets[0].Title != "Go Web Server" {
		t.Errorf("unexpected language filter result: %v", goSnippets[0].Title)
	}
}

func TestGetDefaultSnippets(t *testing.T) {
	defaults := GetDefaultSnippets()
	if len(defaults) == 0 {
		t.Fatal("expected non-empty default snippets")
	}

	for _, snip := range defaults {
		if snip.ID == "" || snip.Title == "" || snip.Language == "" || snip.Code == "" {
			t.Errorf("default snippet missing required fields: %+v", snip)
		}
	}
}
