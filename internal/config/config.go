package config

import "flag"

type Config struct {
	Listen string
	Base   string
}

func (c *Config) New() {
	flag.StringVar(&c.Listen, "a", "127.0.0.1:8888", "Listen on")
	flag.StringVar(&c.Base, "b", "http://127.0.0.1:8000/", "Base address")
	flag.Parse()
}
