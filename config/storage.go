package config

// Storage defines bot data storage configuration.
type Storage struct {
	Type `yaml:"type"`

	FS   `yaml:"fs"`
	Bolt `yaml:"bolt"`
}

// Type defines supported storage types.
type Type string

const (
	TypeFS   Type = "fs"
	TypeBolt Type = "bolt"
)

// FS configuration for file-based storage.
type FS struct {
	Dir string `yaml:"dir"`
}

// Bolt configuration for BoltDB storage.
type Bolt struct {
	Path string `yaml:"path"`
}
