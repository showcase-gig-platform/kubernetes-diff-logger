package config

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"os"
)

func LoadConfig(filename string, cfg *Config) error {
	buf, err := os.ReadFile(filename)
	if err != nil {
		return errors.Wrap(err, "Error reading config file")
	}

	return yaml.UnmarshalStrict(buf, &cfg)
}
