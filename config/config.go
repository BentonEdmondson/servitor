package config

import (
	"errors"
	"fmt"
	"github.com/BurntSushi/toml"
	"os"
	"strings"
	"strconv"
	"time"
)

type Config struct {
	Feeds     map[string][]string    `toml:"feeds"`
	Media	  struct {
		Hook []string `toml:"hook"`
	}	`toml:"media"`
	Style	  struct {
		Colors struct {
			Primary string `toml:"primary"`
			Error string `toml:"error"`
			Highlight string `toml:"highlight"`
			Code string `toml:"code_background"`
		} `toml:"colors"`
	} `toml:"style"`
	Network   struct {
		Context int `toml:"preload_amount"`
		Timeout time.Duration `toml:"timeout_seconds"`
		CacheSize int `toml:"cache_size"`
	} `toml:"network"`
}

var Parsed *Config = nil

/* I use the init function here because everyone who imports config needs
   the config to be parsed before starting, and the config should only be parsed once.
   It seems like a good use case. It is slightly ugly to have printing/exiting
   code this deep in the program, and for it to not be referenced at the top level,
   but ultimately it's not a big deal. */
func init() {
	var err error
	location := location()
	if Parsed, err = parse(location); err != nil {
		os.Stderr.WriteString(fmt.Errorf("failed to parse %s: %w", location, err).Error() + "\n")
		os.Exit(1)
	}
	if err = postprocess(Parsed); err != nil {
		os.Stderr.WriteString(fmt.Errorf("failed to parse %s: %w", location, err).Error() + "\n")
		os.Exit(1)
	}
}

func parse(location string) (*Config, error) {
	/* Default values */
	config := &Config{}
	config.Feeds = map[string][]string{}
	config.Media.Hook = []string{"xdg-open", "%url"}
	config.Style.Context = 5
	config.Style.Colors.Primary = "#A4f59b"
	config.Style.Colors.Error = "#9c3535"
	config.Style.Colors.Highlight = "#0d7d00"
	config.Style.Colors.Code = "#4b4b4b"
	config.Network.Timeout = 10
	config.Network.CacheSize = 128

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
		return nil, fmt.Errorf("contains unrecognized key(s): %v", undecoded)
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

func hexToAnsi(text string) (string, error) {
	errNotAHexCode := errors.New("must be a hex code of the form '#fcba03'")

	if !strings.HasPrefix(text, "#") {
		return "", errNotAHexCode
	}

	if len(text) != 7 {
		return "", errNotAHexCode
	}

	r, err := strconv.ParseUint(text[1:3], 16, 0)
	if err != nil {
		return "", errNotAHexCode
	}
	g, err := strconv.ParseUint(text[3:5], 16, 0)
	if err != nil {
		return "", errNotAHexCode
	}
	b, err := strconv.ParseUint(text[5:7], 16, 0)
	if err != nil {
		return "", errNotAHexCode
	}

	return strconv.Itoa(int(r)) + ";" + strconv.Itoa(int(g)) + ";" + strconv.Itoa(int(b)), nil
}

func postprocess(config *Config) error {
	var err error
	config.Style.Colors.Primary, err = hexToAnsi(config.Style.Colors.Primary)
	if err != nil {
		return fmt.Errorf("key style.colors.primary is invalid: %w", err)
	}
	config.Style.Colors.Error, err = hexToAnsi(config.Style.Colors.Error)
	if err != nil {
		return fmt.Errorf("key style.colors.error is invalid: %w", err)
	}
	config.Style.Colors.Highlight, err = hexToAnsi(config.Style.Colors.Highlight)
	if err != nil {
		return fmt.Errorf("key style.colors.highlight is invalid: %w", err)
	}
	config.Style.Colors.Code, err = hexToAnsi(config.Style.Colors.Code)
	if err != nil {
		return fmt.Errorf("key style.colors.code is invalid: %w", err)
	}
	config.Network.Timeout *= time.Second
	return nil
}
