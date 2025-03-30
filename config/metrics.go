package config

type Metrics struct {
	Enabled bool   `yaml:"enabled"`
	Address string `yaml:"address"`
	Prefix  string `yaml:"prefix"`
}
