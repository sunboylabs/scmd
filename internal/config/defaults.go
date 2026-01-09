package config

import (
	"path/filepath"
)

// Default returns default configuration
func Default() *Config {
	return &Config{
		Version: "1.0",
		Backends: BackendsConfig{
			Default: "llamacpp",
			Local: LocalBackendConfig{
				Model:         "qwen2.5-coder-1.5b",
				ContextLength: 0, // 0 = use model's native context size
				GPULayers:     0,
				Threads:       0,
			},
		},
		UI: UIConfig{
			Streaming: true,
			Colors:    true,
			Verbose:   false,
		},
		Models: ModelsConfig{
			Directory:    filepath.Join(DataDir(), "models"),
			AutoDownload: true,
		},
	}
}
