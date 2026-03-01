// Package config provides configuration management for the URL shortener service.
// Default -> JSON config file -> Command-line flags -> Environment variables
package config

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/caarlos0/env/v11"
	"go.uber.org/zap"
)

// pprofAddr is the address for the pprof server.
const (
	PprofAddr = ":9090"
)

// Config holds the configuration settings for the URL shortener service.
type Config struct {
	Listen          string `env:"SERVER_ADDRESS"`
	Base            string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DataBaseDSN     string `env:"DATABASE_DSN"`
	AuditFile       string `env:"AUDIT_FILE"`
	AuditURL        string `env:"AUDIT_URL"`
	EnableHTTPS     bool   `env:"ENABLE_HTTPS" envDefault:"false"`
	ConfigJSONPath  string `env:"CONFIG"`
	TrustedSubnet   string `env:"TRUSTED_SUBNET"`
}

// JSONConfig is a struct for parsing JSON configuration files.
type ConfigJSON struct {
	Listen          string `json:"server_address"`
	Base            string `json:"base_url"`
	FileStoragePath string `json:"file_storage_path"`
	DataBaseDSN     string `json:"database_dsn"`
	EnableHTTPS     bool   `json:"enable_https"`
	TrustedSubnet   string `json:"trusted_subnet"`
}

// NewDefault initializes a Default Config
func NewDefault() *Config {
	c := &Config{}
	c.Listen = ":8080"
	c.Base = "http://127.0.0.1:8080/"
	return c
}

// New initializes a Config struct with command-line flags.
func New() *Config {
	c := &Config{}
	flag.StringVar(&c.Listen, "a", "", "Listen on")
	flag.StringVar(&c.Base, "b", "", "Base address")
	flag.StringVar(&c.FileStoragePath, "f", "", "File storage path")
	flag.StringVar(&c.DataBaseDSN, "d", "", "Data Base DSN")
	flag.StringVar(&c.AuditFile, "audit-file", "", "Audit log file path")
	flag.StringVar(&c.AuditURL, "audit-url", "", "Audit log ULR")
	flag.BoolVar(&c.EnableHTTPS, "s", false, "Enable HTTPS")
	flag.StringVar(&c.ConfigJSONPath, "c", "", "JSON config file path")
	flag.StringVar(&c.TrustedSubnet, "t", "", "Trusted subnet CIDR notation")
	return c
}

// NewEnv parses environment variables to populate a Config struct.
func NewEnv() *Config {
	c := &Config{}
	err := env.Parse(c)
	if err != nil {
		log.Fatal(err)
	}
	return c
}

// NewConfigJSON initializes a ConfigJSON struct
func NewConfigJSON() *ConfigJSON {
	return &ConfigJSON{}
}

// ApplyEnvOverrides copies non-empty fields from envCfg into cfg.
func ApplyEnvOverrides(cfg *Config, envCfg *Config) {
	defCfg := NewDefault()

	if envCfg.ConfigJSONPath != "" {
		cfg.ConfigJSONPath = envCfg.ConfigJSONPath
	}

	if cfg.ConfigJSONPath != "" {
		cfgJSON := NewConfigJSON()
		err := getConfigJSON(cfg.ConfigJSONPath, cfgJSON)
		if err != nil {
			log.Printf("Failed to load JSON config: %v.\n", err)
		}
		if cfg.Listen == "" {
			cfg.Listen = cfgJSON.Listen
		}
		if cfg.Base == "" {
			cfg.Base = cfgJSON.Base
		}
		if cfg.FileStoragePath == "" {
			cfg.FileStoragePath = cfgJSON.FileStoragePath
		}
		if cfg.DataBaseDSN == "" {
			cfg.DataBaseDSN = cfgJSON.DataBaseDSN
		}
		if !cfg.EnableHTTPS {
			cfg.EnableHTTPS = cfgJSON.EnableHTTPS
		}
		if cfg.TrustedSubnet == "" {
			cfg.TrustedSubnet = cfgJSON.TrustedSubnet
		}
	}

	if envCfg.Listen != "" {
		cfg.Listen = envCfg.Listen
	}
	if envCfg.Base != "" {
		cfg.Base = envCfg.Base
	}
	if envCfg.DataBaseDSN != "" {
		cfg.DataBaseDSN = envCfg.DataBaseDSN
	}
	if envCfg.FileStoragePath != "" {
		cfg.FileStoragePath = envCfg.FileStoragePath
	}
	if envCfg.AuditFile != "" {
		cfg.AuditFile = envCfg.AuditFile
	}
	if envCfg.AuditURL != "" {
		cfg.AuditURL = envCfg.AuditURL
	}
	if envCfg.EnableHTTPS {
		cfg.EnableHTTPS = envCfg.EnableHTTPS
	}

	if cfg.Listen == "" {
		cfg.Listen = defCfg.Listen
	}
	if cfg.Base == "" {
		cfg.Base = defCfg.Base
	}
	if cfg.TrustedSubnet == "" {
		cfg.TrustedSubnet = defCfg.TrustedSubnet
	}
}

// NewLogger creates and returns a SugaredLogger for logging.
func NewLogger() *zap.SugaredLogger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	return logger.Sugar()
}

// open loads the data from the JSON file into memory.
func getConfigJSON(filename string, configJSON *ConfigJSON) error {
	file, err := os.ReadFile(filename)
	if err != nil {
		return err
	} else {
		err = json.Unmarshal(file, &configJSON)
		if err != nil {
			return err
		}
		return nil
	}
}
