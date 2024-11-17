package v1

import (
	"gopkg.in/yaml.v3"
	"os"
)

var GLOBAL_EXCLUDE = []string{"node_modules", ".git", ".nuxt", "test"}

type NfsWatchConfig struct {
	Pattern string
	Hooks   []string
}

type NfsPodConfig struct {
	Cwd       string `yaml:"cwd"`
	Selector  string `yaml:"selector"`
	Namespace string `yaml:"namespace"`
}

type NfsConfig struct {
	Manifest    string           `yaml:"manifest"`
	PodConfig   NfsPodConfig     `yaml:"pod"`
	WatchConfig []NfsWatchConfig `yaml:"watch"`
	Paralell    bool             `yaml:"paralell"`
	Interval    uint32           `yaml:"interval"`
}

func Parse() NfsConfig {
	file, err := os.ReadFile(".nfs.yml")
	if err != nil {
		panic(err)
	}

	// Create default and overwrite what user configured
	config := NfsConfig{Paralell: false, Interval: 1000}

	err = yaml.Unmarshal(file, &config)

	if err != nil {
		panic(err)
	}

	return config
}
