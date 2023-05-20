package http

import "time"

type ServerConfig struct {
	Address      string        `yaml:"address"`
	ReadTimeout  time.Duration `yaml:"readtimeout"`
	WriteTimeout time.Duration `yaml:"writetimeout"`
}
