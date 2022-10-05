package config

type Config struct {
	Differs                []DifferConfig `yaml:"differs"`
	CommonLabelConfig      ExtraConfig    `yaml:"commonLabelConfig"`
	CommonAnnotationConfig ExtraConfig    `yaml:"commonAnnotationConfig"`
}

type DifferConfig struct {
	GroupKind  GroupKind `yaml:"groupKind"`
	NameFilter string    `yaml:"nameFilter"`
}

type GroupKind struct {
	Group string `yaml:"group"`
	Kind  string `yaml:"kind"`
}

type ExtraConfig struct {
	Enable     bool     `yaml:"enable"`
	IgnoreKeys []string `yaml:"ignoreKeys"`
}
