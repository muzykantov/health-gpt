package config

// Telegram defines Telegram bot configuration.
type Telegram struct {
	Token string `yaml:"token"`
	Debug bool   `yaml:"debug"`
}
