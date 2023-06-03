package config

import (
	"fmt"
	"errors"
	"os"
	"github.com/BurntSushi/toml"
)

type Config struct {
	General general
	Feeds feeds
	Algos algos
}

type general = struct {
	Context int
	Timeout int
}
type feeds = map[string][]string
type algos = map[string]struct {
	Server string
	Query string
}

func Parse() (*Config, error) {
	config := &Config {
		General: general{
			Context: 5,
			Timeout: 10,
		},
		Feeds: feeds{},
		Algos: algos{},
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
		return nil, fmt.Errorf("config file %s contained unexpected keys: %v", location, undecoded)
	}

	return config, nil
}

func location() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return xdg + "/mimicry/config.toml"
	}

	if home := os.Getenv("HOME"); home != "" {
		return home + "/.config/mimicry/config.toml"
	}

	return ""
}
