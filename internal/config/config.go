package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/oolong-sh/sync"
)

var cfg OolongConfig = defaultConfig()

type OolongSyncConfig sync.SyncConfig

type OolongConfig struct {
	NotesDirPaths     []string `toml:"note_directories"`
	AllowedExtensions []string `toml:"allowed_extensions"`
	IgnoreDirectories []string `toml:"ignored_directories"`
	OpenCommand       []string `toml:"open_command"`
	PinningEnabled    bool     `toml:"pinning_enabled"`

	LinkerConfig  OolongLinkerConfig `toml:"linker"`
	GraphConfig   OolongGraphConfig  `toml:"graph"`
	SyncConfig    OolongSyncConfig   `toml:"sync"`
	PluginsConfig OolongPluginConfig `toml:"plugins"`
}

type OolongLinkerConfig struct {
	NGramRange []int    `toml:"ngram_range"`
	StopWords  []string `toml:"stop_words"`
}

type OolongGraphConfig struct {
	MinNodeWeight float64 `toml:"min_node_weight"`
	MaxNodeWeight float64 `toml:"max_node_weight"`
	MinLinkWeight float64 `toml:"min_link_weight"`
	DefaultMode   string  `toml:"default_mode"`
}

type OolongEditorConfig struct {
	// TODO: web editor related config (themes?)
}

type OolongPluginConfig struct {
	PluginPaths []string `toml:"plugin_paths"`
}

func Config() *OolongConfig { return &cfg }

func NotesDirPaths() []string             { return cfg.NotesDirPaths }
func OpenCommand() []string               { return cfg.OpenCommand }
func AllowedExtensions() []string         { return cfg.AllowedExtensions }
func PluginPaths() []string               { return cfg.PluginsConfig.PluginPaths }
func IgnoredDirectories() []string        { return cfg.IgnoreDirectories }
func LinkerConfig() OolongLinkerConfig    { return cfg.LinkerConfig }
func NGramRange() []int                   { return cfg.LinkerConfig.NGramRange }
func StopWords() []string                 { return cfg.LinkerConfig.StopWords }
func WeightThresholds() OolongGraphConfig { return cfg.GraphConfig }
func GraphMode() string                   { return cfg.GraphConfig.DefaultMode }
func SyncConfig() OolongSyncConfig        { return cfg.SyncConfig }
func PinningEnabled() bool                { return cfg.PinningEnabled }

func Setup() error {
	configPath, err := findConfigPath()
	if err != nil {
		panic(err)
	}

	readConfig(configPath)

	go initWatcher(configPath)

	return nil
}

func readConfig(configPath string) {
	fmt.Println("Parsing configuration:", configPath)

	contents, err := os.ReadFile(configPath)
	if err != nil {
		panic(err)
	}

	err = toml.Unmarshal(contents, &cfg)
	if err != nil {
		panic(err)
	}

	for i, dir := range cfg.NotesDirPaths {
		d, err := expand(dir)
		if err != nil {
			panic(err)
		}
		cfg.NotesDirPaths[i] = d
	}
}
