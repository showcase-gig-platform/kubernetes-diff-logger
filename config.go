package main

// Config represents the configuration options for kubernetes-diff-logger
type Config struct {
	GroupKinds []GroupKindConfig `yaml:"groupKinds"`
}

type GroupKindConfig struct {
	Group      string `yaml:"group"`
	Kind       string `yaml:"kind"`
	NameFilter string `yaml:"nameFilter"`
}

// DefaultConfig returns a default deployment watching config
func DefaultConfig() Config {
	return Config{
		GroupKinds: []GroupKindConfig{
			GroupKindConfig{
				NameFilter: "*",
				Group:      "apps",
				Kind:       "Deployment",
			},
		},
	}
}
