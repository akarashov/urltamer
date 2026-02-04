// Package config provides configuration management for the URL shortener service.
package config

import (
	"flag"
	"log"

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
}

// New initializes a Config struct with command-line flags.
func New() *Config {
	c := &Config{}
	flag.StringVar(&c.Listen, "a", ":8080", "Listen on")
	flag.StringVar(&c.Base, "b", "http://127.0.0.1:8080/", "Base address")
	flag.StringVar(&c.FileStoragePath, "f", "", "File storage path")
	flag.StringVar(&c.DataBaseDSN, "d", "", "Data Base DSN")
	flag.StringVar(&c.AuditFile, "audit-file", "", "Audit log file path")
	flag.StringVar(&c.AuditURL, "audit-url", "", "Audit log ULR")
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

// NewLogger creates and returns a SugaredLogger for logging.
func NewLogger() *zap.SugaredLogger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	return logger.Sugar()
}
