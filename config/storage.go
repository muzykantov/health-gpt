package config

// Storage defines bot data storage configuration.
type Storage struct {
	Type `yaml:"type"`

	Filesystem `yaml:"filesystem"`
	Redis      `yaml:"redis"`
}

// Type defines supported storage types.
type Type string

const (
	TypeFS    Type = "fs"
	TypeRedis Type = "redis"
)

// Filesystem configuration for file-based storage.
type Filesystem struct {
	Path string `yaml:"path"`
}

// Redis configuration for Redis storage.
type Redis struct {
	Address    string `yaml:"address"`
	Password   string `yaml:"password"`
	DB         int    `yaml:"db"`
	Expiration string `yaml:"expiration"`
}
