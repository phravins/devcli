package assets

import (
	"testing"
)

func TestGetLogo(t *testing.T) {
	logo, err := GetLogo()
	if err != nil {
		t.Fatalf("Failed to get logo: %v", err)
	}

	if len(logo) == 0 {
		t.Error("Logo is empty")
	}

	// PNG files should start with PNG magic number: 89 50 4E 47
	if len(logo) < 4 {
		t.Error("Logo too small to be a valid PNG")
	}
	if logo[0] != 0x89 || logo[1] != 0x50 || logo[2] != 0x4E || logo[3] != 0x47 {
		t.Error("Logo does not have PNG magic number")
	}
}

func TestGetAsset(t *testing.T) {
	// Test getting logo via generic GetAsset function
	logo, err := GetAsset("devcli_logo.png")
	if err != nil {
		t.Fatalf("Failed to get asset: %v", err)
	}

	if len(logo) == 0 {
		t.Error("Asset is empty")
	}
}

func TestAssetExists(t *testing.T) {
	tests := []struct {
		name     string
		asset    string
		expected bool
	}{
		{"Logo exists", "devcli_logo.png", true},
		{"Non-existent file", "nonexistent.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists := AssetExists(tt.asset)
			if exists != tt.expected {
				t.Errorf("AssetExists(%q) = %v, want %v", tt.asset, exists, tt.expected)
			}
		})
	}
}

func TestListAssets(t *testing.T) {
	assets, err := ListAssets()
	if err != nil {
		t.Fatalf("Failed to list assets: %v", err)
	}

	if len(assets) == 0 {
		t.Error("No assets found")
	}

	// Check that devcli_logo.png is in the list
	found := false
	for _, asset := range assets {
		if asset == "devcli_logo.png" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("devcli_logo.png not found in assets list: %v", assets)
	}
}

func TestGetAssetsFS(t *testing.T) {
	fs := GetAssetsFS()
	if fs == nil {
		t.Fatal("GetAssetsFS returned nil")
	}

	// Try to open the logo file
	file, err := fs.Open("devcli_logo.png")
	if err != nil {
		t.Fatalf("Failed to open logo from FS: %v", err)
	}
	defer file.Close()

	// Read some bytes to verify it's accessible
	buf := make([]byte, 4)
	n, err := file.Read(buf)
	if err != nil {
		t.Fatalf("Failed to read from file: %v", err)
	}
	if n != 4 {
		t.Errorf("Expected to read 4 bytes, got %d", n)
	}
}
