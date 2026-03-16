package mcp

import (
	"os"

	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	Name    string   `json:"name" yaml:"name"`
	Mode    string   `json:"mode" yaml:"mode"`
	Command string   `json:"command,omitempty" yaml:"command,omitempty"`
	Args    []string `json:"args,omitempty" yaml:"args,omitempty"`
	URL     string   `json:"url,omitempty" yaml:"url,omitempty"`
}

type ConfigFile struct {
	Servers []ServerConfig `yaml:"mcp_servers"`
}

func LoadConfig(path string) ([]ServerConfig, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var cfg ConfigFile
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}
	return cfg.Servers, nil
}
