package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	HTTP *HTTPConfig     `json:"http"`
	Log  *LogConfig      `json:"log"`
	DB   *DatabaseConfig `json:"db"`
}

func LoadConfig(file string) *Config {
	yamlFile, err := os.ReadFile(file)
	if err != nil {
		panic(err)
	}
	var config Config
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		panic(err)
	}
	return &config
}

type HTTPConfig struct {
	Host  string `json:"host"`
	Port  string `json:"port"`
	Pprof bool   `json:"pprof"`
}

type LogConfig struct {
	Level string `json:"level"`
	Path  string `json:"path"`
}

type DatabaseConfig struct {
	Name string `json:"name"`
}
