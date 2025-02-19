package config

// Storage defines bot data storage configuration.
type Storage struct {
	Type `yaml:"type"`

	PostgreSQL `yaml:"postgresql"`
	Filesystem `yaml:"filesystem"`
}

// Type defines supported storage types.
type Type string

const (
	TypePostgreSQL Type = "postgresql"
	TypeFS         Type = "fs"
	TypeInMemory   Type = "in-memory"
)

// PostgreSQL configuration for PostgreSQL storage.
type PostgreSQL struct {
	Host      string `yaml:"host"`
	Port      int    `yaml:"port"`
	User      string `yaml:"user"`
	Password  string `yaml:"password"`
	Database  string `yaml:"database"`
	SSLMode   string `yaml:"ssl_mode"`
	Migration bool   `yaml:"migration"`
}

// Filesystem configuration for file-based storage.
type Filesystem struct {
	Path string `yaml:"path"`
}
