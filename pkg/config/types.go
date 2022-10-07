package config

type Config struct {
	Differs                []DifferConfig `yaml:"differs"`
	CommonLabelConfig      ExtraConfig    `yaml:"commonLabelConfig"`
	CommonAnnotationConfig ExtraConfig    `yaml:"commonAnnotationConfig"`
}

type DifferConfig struct {
	GroupKind    GroupKind `yaml:"groupKind"`
	MatchRegexp  string    `yaml:"matchRegexp"`
	IgnoreRegexp string    `yaml:"ignoreRegexp"`
}

type GroupKind struct {
	Group string `yaml:"group"`
	Kind  string `yaml:"kind"`
}

type ExtraConfig struct {
	Enable     bool     `yaml:"enable"`
	IgnoreKeys []string `yaml:"ignoreKeys"`
}
