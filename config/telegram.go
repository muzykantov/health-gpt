package config

// Telegram defines Telegram bot configuration.
type Telegram struct {
	Token string `yaml:"token" validate:"required,len=46"`
	Debug bool   `yaml:"debug"`
}
