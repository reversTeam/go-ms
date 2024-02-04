package core

import (
	"os"

	"sigs.k8s.io/yaml"
)

type Config struct {
	Grpc     *ServerConfig            `yaml:"grpc,omitempty"`
	Http     *HttpServerConfig        `yaml:"http,omitempty"`
	Exporter ExporterConfig           `yaml:"exporter,omitempty"`
	Services map[string]ServiceConfig `yaml:"services"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type HttpServerConfig struct {
	ServerConfig
	ReadTimeout  int `yaml:"read-timeout"`
	WriteTimeout int `yaml:"write-timeout"`
}

type ExporterConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Path     string `yaml:"path"`
	Interval int    `yaml:"interval"`
}

type ServiceConfig struct {
	Grpc   bool                   `yaml:"grpc"`
	Http   bool                   `yaml:"http"`
	Config map[string]interface{} `yaml:"config,omitempty"`
}

func NewConfig(filename string) (*Config, error) {
	buf, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(buf, &config)
	if err != nil {
		return nil, err
	}

	// Set default values for HttpServerConfig if they are not provided
	if config.Http != nil {
		if config.Http.ReadTimeout == 0 {
			config.Http.ReadTimeout = 60
		}
		if config.Http.WriteTimeout == 0 {
			config.Http.WriteTimeout = 60
		}
	}

	return &config, nil
}
