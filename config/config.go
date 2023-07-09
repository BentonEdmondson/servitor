package config

import (
	"errors"
	"fmt"
	"github.com/BurntSushi/toml"
	"os"
)

type Config struct {
	Context   int      `toml:"context"`
	Timeout   int      `toml:"timeout"`
	Feeds     feeds    `toml:"feeds"`
	Algos     algos    `toml:"algos"`
	MediaHook []string `toml:"media_hook"`
}

type feeds = map[string][]string
type algos = map[string]struct {
	Server string `toml:"server"`
	Query  string `toml:"query"`
}

func Parse() (*Config, error) {
	/* Default values */
	config := &Config{
		Context:   5,
		Timeout:   10,
		Feeds:     feeds{},
		Algos:     algos{},
		MediaHook: []string{"xdg-open", "%url"},
	}

	location := location()
	if location == "" {
		return config, nil
	}

	metadata, err := toml.DecodeFile(location, config)
	if errors.Is(err, os.ErrNotExist) {
		return config, nil
	}
	if err != nil {
		return nil, err
	}

	if undecoded := metadata.Undecoded(); len(undecoded) != 0 {
		return nil, fmt.Errorf("config file %s contained unrecognized keys: %v", location, undecoded)
	}

	return config, nil
}

func location() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return xdg + "/servitor/config.toml"
	}

	if home := os.Getenv("HOME"); home != "" {
		return home + "/.config/servitor/config.toml"
	}

	return ""
}
