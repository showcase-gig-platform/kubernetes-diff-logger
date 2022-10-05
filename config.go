package main

import "github.com/showcase-gig-platform/kubernetes-diff-logger/pkg/differ"

// Config represents the configuration options for kubernetes-diff-logger
type Config struct {
	Differs                []DifferConfig     `yaml:"differs"`
	CommonLabelConfig      differ.ExtraConfig `yaml:"commonLabelConfig"`
	CommonAnnotationConfig differ.ExtraConfig `yaml:"commonAnnotationConfig"`
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
			{
				NameFilter: "*",
				GroupKind: GroupKind{
					Group: "apps",
					Kind:  "deployment",
				},
			},
		},
		CommonLabelConfig: differ.ExtraConfig{
			Enable:     false,
			IgnoreKeys: []string{},
		},
		CommonAnnotationConfig: differ.ExtraConfig{
			Enable:     false,
			IgnoreKeys: []string{},
		},
	}
}
