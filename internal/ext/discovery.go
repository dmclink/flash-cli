package ext

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/dmclink/flash-cli/internal/config"
	"github.com/dmclink/flash-cli/shared"
)

var checkCapabilityMap = map[string]func(PluginManifest) bool{
	shared.CAPABILITY_RENDER:           func(m PluginManifest) bool { return m.Capabilities.Renderer },
	shared.CAPABILITY_REVIEW_PROCESSOR: func(m PluginManifest) bool { return m.Capabilities.ReviewProcessor },
}

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
func FindPlugin(name string, capabilityName string) (*PluginManifest, string, error) {
	checkCapability, ok := checkCapabilityMap[capabilityName]
	if !ok {
		return nil, "", fmt.Errorf("unexpected capability name: %s", capabilityName)
	}

	pluginsDir := config.V.GetString(config.KeyPathPluginsDir)
	fmt.Println("pluginsDir: ", pluginsDir)

	dirs, err := os.ReadDir(pluginsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, "", nil
		}
		return nil, "", fmt.Errorf("reading plugin dirs at '%s' | %w", pluginsDir, err)
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
