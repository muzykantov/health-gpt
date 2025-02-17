package config

// Storage defines bot data storage configuration.
type Storage struct {
	Type `yaml:"type" validate:"required,oneof=postgresql fs in-memory"`

	Postgresql `yaml:"postgresql" validate:"required_if=Type postgresql"`
	Filesystem `yaml:"filesystem" validate:"required_if=Type fs"`
}

// Type defines supported storage types.
type Type string

const (
	TypePostgreSQL Type = "postgresql"
	TypeFS         Type = "fs"
	TypeInMemory   Type = "in-memory"
)

// Postgresql configuration for PostgreSQL storage.
type Postgresql struct {
	Host     string `yaml:"host" validate:"required"`
	Port     int    `yaml:"port" validate:"required,min=1,max=65535"`
	User     string `yaml:"user" validate:"required"`
	Password string `yaml:"password" validate:"required"`
	Database string `yaml:"database" validate:"required"`
	SSLMode  string `yaml:"ssl_mode" validate:"omitempty,oneof=disable require verify-ca verify-full"`
}

// Filesystem configuration for file-based storage.
type Filesystem struct {
	Path string `yaml:"path" validate:"required,dir"`
}
