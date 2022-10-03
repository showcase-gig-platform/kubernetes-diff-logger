package main

// Config represents the configuration options for kubernetes-diff-logger
type Config struct {
	Differs []DifferConfig `yaml:"differs"`
}

type DifferConfig struct {
	GroupKind  GroupKind `yaml:"groupKind"`
	NameFilter string    `yaml:"nameFilter"`
}

type GroupKind struct {
	Group string `yaml:"group"`
	Kind  string `yaml:"kind"`
}

// DefaultConfig returns a default deployment watching config
func DefaultConfig() Config {
	return Config{
		Differs: []DifferConfig{
			DifferConfig{
				NameFilter: "*",
				GroupKind: GroupKind{
					Group: "apps",
					Kind:  "deployment",
				},
			},
		},
	}
}
