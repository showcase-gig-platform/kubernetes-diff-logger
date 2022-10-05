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
