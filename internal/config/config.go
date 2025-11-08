package config

import (
	"flag"
	"log"
	"github.com/caarlos0/env/v11"
	"go.uber.org/zap"
)

type Config struct {
	Listen string `env:"SERVER_ADDRESS"`
	Base   string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DataBaseDSN string `env:"DATABASE_DSN"`
}

func New() *Config {
	c := &Config{}
	flag.StringVar(&c.Listen, "a", ":8080", "Listen on")
	flag.StringVar(&c.Base, "b", "http://127.0.0.1:8080/", "Base address")
	flag.StringVar(&c.FileStoragePath, "f", "./tamers.json", "File storage path")
	flag.StringVar(&c.DataBaseDSN, "d", "postgres://admin:admin@192.168.0.20:5432/demo?sslmode=disable", "Data Base DSN")
	return c
}

func NewEnv() *Config {
	c := &Config{}
	err := env.Parse(c)
	if err != nil {
		log.Fatal(err)
	}
	return c
}

func NewLogger() *zap.SugaredLogger{
	logger, err := zap.NewDevelopment()
    if err != nil {
        panic(err)
    }
    defer logger.Sync()
    return logger.Sugar()
}