
package assets

import (
	"embed"
	"io/fs"
)
var assetsFS embed.FS
func GetLogo() ([]byte, error) {
	return assetsFS.ReadFile("devcli_logo.png")
}
func GetAsset(name string) ([]byte, error) {
	return assetsFS.ReadFile(name)
}
func GetAssetsFS() fs.FS {
	return assetsFS
}
func ListAssets() ([]string, error) {
	var files []string
	err := fs.WalkDir(assetsFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && path != "." {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}
func AssetExists(name string) bool {
	_, err := assetsFS.ReadFile(name)
	return err == nil
}
