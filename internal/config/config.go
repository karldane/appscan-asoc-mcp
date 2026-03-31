package config

import (
	"flag"
	"os"
	"strconv"
	"strings"
	"sync"
)

var initFlags sync.Once

func ResetForTest() {
	initFlags = sync.Once{}
}

type Config struct {
	BaseURL        string
	KeyID          string
	KeySecret      string
	TimeoutSeconds int
	isReadOnly     bool
	LogJSON        bool
}

func (c *Config) ReadOnly() bool {
	return c.isReadOnly
}

func Load() *Config {
	initFlags.Do(func() {
		flag.Parse()
	})

	cfg := &Config{
		TimeoutSeconds: 30,
		isReadOnly:     false,
		LogJSON:        false,
	}

	if v := os.Getenv("APPSCAN_BASE_URL"); v != "" {
		cfg.BaseURL = v
	}

	if v := os.Getenv("APPSCAN_API_KEY"); v != "" {
		parts := strings.SplitN(v, ":", 2)
		if len(parts) == 2 {
			cfg.KeyID = parts[0]
			cfg.KeySecret = parts[1]
		}
	}

	if cfg.KeyID == "" {
		if v := os.Getenv("APPSCAN_KEY_ID"); v != "" {
			cfg.KeyID = v
		}
	}
	if cfg.KeySecret == "" {
		if v := os.Getenv("APPSCAN_KEY_SECRET"); v != "" {
			cfg.KeySecret = v
		}
	}

	if v := os.Getenv("APPSCAN_TIMEOUT_SECONDS"); v != "" {
		if secs, err := strconv.Atoi(v); err == nil && secs > 0 {
			cfg.TimeoutSeconds = secs
		}
	}

	return cfg
}

func MergeCLIFlags(cfg *Config) *Config {
	if flag.Lookup("base-url") != nil {
		if v := flag.Lookup("base-url").Value.String(); v != "" {
			cfg.BaseURL = v
		}
	}
	if flag.Lookup("timeout") != nil {
		if v := flag.Lookup("timeout").Value.String(); v != "" {
			if secs, err := strconv.Atoi(v); err == nil && secs > 0 {
				cfg.TimeoutSeconds = secs
			}
		}
	}
	if flag.Lookup("readonly") != nil {
		cfg.isReadOnly = flag.Lookup("readonly").Value.String() == "true"
	}
	if flag.Lookup("log-json") != nil {
		cfg.LogJSON = flag.Lookup("log-json").Value.String() == "true"
	}
	return cfg
}

func init() {
	flag.String("base-url", "", "Override APPSCAN_BASE_URL")
	flag.String("timeout", "", "Override HTTP timeout in seconds")
	flag.Bool("readonly", false, "Disable all mutating tools")
	flag.Bool("log-json", false, "Emit structured JSON logs")
}
