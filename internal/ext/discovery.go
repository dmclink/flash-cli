package ext

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/dmclink/flash-cli/internal/constant"
)

// PluginManifest represents the structured data required inside every plugin.toml file
type PluginManifest struct {
	Name        string `toml:"name"`
	Alias       string `toml:"alias,omitempty"`
	Version     string `toml:"version"`
	Author      string `toml:"author"`
	Description string `toml:"description"`
	Executable  string `toml:"executable"`

	Capabilities struct {
		ReviewProcessor bool `toml:"review_processor"`
		Renderer        bool `toml:"renderer"`
	} `toml:"capabilities"`
}

// FindPlugin handles the core logic and filters plugins based on a custom capability check.
func FindPlugin(name string, checkCapability func(PluginManifest) bool) (*PluginManifest, string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, "", fmt.Errorf("getting user home dir | %w", err)
	}

	pluginsDir := filepath.Join(home, ".config", constant.APP_NAME, "plugins")
	dirs, err := os.ReadDir(pluginsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, "", nil
		}
		return nil, "", fmt.Errorf("reading plugin dirs | %w", err)
	}

	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}

		manifestPath := filepath.Join(pluginsDir, dir.Name(), "plugin.toml")
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			fmt.Printf("Error: error reading or missing manifest file at '%s'. Skipping.\n", manifestPath)
			continue
		}

		var manifest PluginManifest
		meta, err := toml.Decode(string(data), &manifest)
		if err != nil {
			fmt.Printf("Error: malformed manifest; not valid toml at '%s'. Skipping.\n", manifestPath)
			continue
		}

		if undecoded := meta.Undecoded(); len(undecoded) > 0 {
			fmt.Printf("Warning: Plugin at '%s' has unknown fields: %v.\n", manifestPath, undecoded)
		}

		if checkCapability(manifest) && (manifest.Alias == name || manifest.Name == name) {
			binaryPath := filepath.Join(pluginsDir, dir.Name(), manifest.Executable)
			return &manifest, binaryPath, nil
		}
	}

	return nil, "", fmt.Errorf("plugin '%s' not found with required capabilities", name)
}

// FindReviewPlugin finds the first plugin with ReviewProcessor capability and matching name
func FindReviewPlugin(name string) (*PluginManifest, string, error) {
	return FindPlugin(name, func(m PluginManifest) bool {
		return m.Capabilities.ReviewProcessor
	})
}

// FindRendererPlugin finds the first plugin plugins with Renderer capability and matching name
func FindRendererPlugin(name string) (*PluginManifest, string, error) {
	return FindPlugin(name, func(m PluginManifest) bool {
		return m.Capabilities.Renderer
	})
}
