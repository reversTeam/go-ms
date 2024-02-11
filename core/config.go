package core

import (
	"os"

	"sigs.k8s.io/yaml"
)

type Config struct {
	Name       string                   `yaml:"name"`
	Grpc       *ServerConfig            `yaml:"grpc,omitempty"`
	Http       *HttpServerConfig        `yaml:"http,omitempty"`
	Exporter   ExporterConfig           `yaml:"exporter,omitempty"`
	Jaeger     JaegerConfig             `yaml:"jaeger,omitempty"`
	Services   map[string]ServiceConfig `yaml:"services"`
	Middleware []string                 `yaml:"middlewares,omitempty"`
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
	Http   bool                   `yaml:"http"`
	Config map[string]interface{} `yaml:"config,omitempty"`
}

type JaegerConfig struct {
	Name     string `yaml:"name"`
	Unsecure bool   `yaml:"unsecure,omitempty"`
	Mode     string `yaml:"mode"`
	Host     string `yaml:"host,omitempty"`
	Port     int    `yaml:"port,omitempty"`
	Sampler  struct {
		Type  string `yaml:"type,omitempty"`
		Param int    `yaml:"param,omitempty"`
	} `yaml:"sampler,omitempty"`
	Reporter struct {
		LogSpans bool `yaml:"log-spans,omitempty"`
	} `yaml:"reporter,omitempty"`
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
