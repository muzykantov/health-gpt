package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
)

// Configuration structure hierarchy represents a tree where each node
// can be configured via both JSON and environment variables.
// Order of precedence: environment variables > JSON > default values

// BotConfig represents the root configuration of the application.
// Each subsystem (LLM, Storage, etc.) has its own configuration section
// to maintain separation of concerns.
type BotConfig struct {
	// LLM section contains settings for language model providers
	LLM LLMConfig
	// Telegram section contains bot API credentials and settings
	Telegram TelegramConfig
	// Storage section defines how different types of data are stored
	Storage StorageConfig
}

// LLMConfig defines settings for language model interactions.
// Supports multiple providers through a provider-specific configuration.
type LLMConfig struct {
	// ActiveProvider determines which provider configuration to use
	// Valid values: "gigachat", "chatgpt"
	ActiveProvider string
	// Configuration sections for each supported provider
	GigaChat GigaChatConfig
	ChatGPT  ChatGPTConfig
}

// GigaChatConfig contains settings specific to GigaChat provider.
// All floating-point parameters use float64 for consistency.
type GigaChatConfig struct {
	// Authentication credentials
	ClientID     string
	ClientSecret string
	// Model configuration
	Model             string  `json:"model"`              // Default: "GigaChat-Max"
	Temperature       float64 `json:"temperature"`        // Range: [0.0, 2.0], Default: 0.1
	TopP              float64 `json:"top_p"`              // Range: [0.0, 1.0], Default: 1.0
	MaxTokens         int     `json:"max_tokens"`         // Default: 2048
	RepetitionPenalty float64 `json:"repetition_penalty"` // Default: 1.0
}

// ChatGPTConfig contains settings specific to ChatGPT provider.
// All floating-point parameters use float64 for consistency.
type ChatGPTConfig struct {
	// Authentication credentials
	APIKey string
	// Model configuration
	Model       string  `json:"model"`       // Default: "gpt-4"
	Temperature float64 `json:"temperature"` // Range: [0.0, 2.0], Default: 0.1
	TopP        float64 `json:"top_p"`       // Range: [0.0, 1.0], Default: 1.0
	MaxTokens   int     `json:"max_tokens"`  // Default: 2048
	SocksProxy  string  `json:"socks_proxy"` // Optional SOCKS proxy URL
}

// TelegramConfig contains settings for the Telegram bot API.
type TelegramConfig struct {
	Token string // Bot API token from BotFather
	Debug bool   // Enables debug logging when true
}

// StorageConfig defines how different types of data are persisted.
// Each data type can use its own storage backend.
type StorageConfig struct {
	// Database configuration for persistent storage
	Postgres PostgresConfig
	// Chat history storage configuration
	History HistoryStorageConfig
	// User data storage configuration
	Users UserStorageConfig
}

// PostgresConfig contains all necessary settings for PostgreSQL connection.
// Supports configuration via environment variables for secure deployment.
type PostgresConfig struct {
	Host     string `json:"host"`     // Database host
	Port     int    `json:"port"`     // Database port
	User     string `json:"user"`     // Database user
	Password string `json:"password"` // Database password
	Database string `json:"database"` // Database name
	SSLMode  string `json:"ssl_mode"` // SSL mode (disable, require, verify-full)
}

// HistoryStorageConfig defines how chat history is stored.
type HistoryStorageConfig struct {
	// Storage type: "postgres", "file", "memory"
	Type string `json:"type"`
	// File path for "file" storage type
	Path string `json:"path"`
}

// UserStorageConfig defines how user data is stored.
type UserStorageConfig struct {
	// Storage type: "postgres", "file", "memory"
	Type string `json:"type"`
	// File path for "file" storage type
	Path string `json:"path"`
}

// LoadConfig reads configuration from reader and applies environment variables.
// Environment variables take precedence over file configuration.
func LoadConfig(r io.Reader) (*BotConfig, error) {
	var cfg BotConfig
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("decode JSON configuration: %w", err)
	}

	applyEnvironmentVariables(&cfg)

	if err := validateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("validate configuration: %w", err)
	}

	setDefaults(&cfg)
	return &cfg, nil
}

// LoadBotConfigFromFile loads configuration from the specified file.
func LoadBotConfigFromFile(filename string) (*BotConfig, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("open config file: %w", err)
	}
	defer file.Close()

	return LoadConfig(file)
}

// applyEnvironmentVariables overrides configuration with environment variables.
func applyEnvironmentVariables(cfg *BotConfig) {
	// LLM Configuration
	if v := os.Getenv("LLM_PROVIDER"); v != "" {
		cfg.LLM.ActiveProvider = v
	}

	// GigaChat Configuration
	if v := os.Getenv("GIGACHAT_CLIENT_ID"); v != "" {
		cfg.LLM.GigaChat.ClientID = v
	}
	if v := os.Getenv("GIGACHAT_CLIENT_SECRET"); v != "" {
		cfg.LLM.GigaChat.ClientSecret = v
	}

	// ChatGPT Configuration
	if v := os.Getenv("CHATGPT_API_KEY"); v != "" {
		cfg.LLM.ChatGPT.APIKey = v
	}

	// Telegram Configuration
	if v := os.Getenv("TELEGRAM_TOKEN"); v != "" {
		cfg.Telegram.Token = v
	}

	// PostgreSQL Configuration
	if v := os.Getenv("POSTGRES_HOST"); v != "" {
		cfg.Storage.Postgres.Host = v
	}
	if v := os.Getenv("POSTGRES_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.Storage.Postgres.Port = port
		}
	}
	if v := os.Getenv("POSTGRES_USER"); v != "" {
		cfg.Storage.Postgres.User = v
	}
	if v := os.Getenv("POSTGRES_PASSWORD"); v != "" {
		cfg.Storage.Postgres.Password = v
	}
	if v := os.Getenv("POSTGRES_DB"); v != "" {
		cfg.Storage.Postgres.Database = v
	}
	if v := os.Getenv("POSTGRES_SSL_MODE"); v != "" {
		cfg.Storage.Postgres.SSLMode = v
	}
}

// validateConfig ensures all required fields are set and values are within valid ranges.
func validateConfig(cfg *BotConfig) error {
	// Validate LLM configuration
	if err := validateLLMConfig(&cfg.LLM); err != nil {
		return fmt.Errorf("LLM config: %w", err)
	}

	// Validate Telegram configuration
	if cfg.Telegram.Token == "" {
		return fmt.Errorf("telegram token is required")
	}

	// Validate PostgreSQL configuration when used
	if cfg.Storage.History.Type == "postgres" || cfg.Storage.Users.Type == "postgres" {
		if err := validatePostgresConfig(&cfg.Storage.Postgres); err != nil {
			return fmt.Errorf("postgreSQL config: %w", err)
		}
	}

	// Validate file paths when file storage is used
	if cfg.Storage.History.Type == "file" && cfg.Storage.History.Path == "" {
		return fmt.Errorf("history storage: path is required for file storage")
	}
	if cfg.Storage.Users.Type == "file" && cfg.Storage.Users.Path == "" {
		return fmt.Errorf("user storage: path is required for file storage")
	}

	return nil
}

// validateLLMConfig validates language model configuration.
func validateLLMConfig(cfg *LLMConfig) error {
	if cfg.ActiveProvider == "" {
		return fmt.Errorf("active provider is required")
	}

	switch cfg.ActiveProvider {
	case "gigachat":
		return validateGigaChatConfig(&cfg.GigaChat)
	case "chatgpt":
		return validateChatGPTConfig(&cfg.ChatGPT)
	default:
		return fmt.Errorf("unknown provider: %s", cfg.ActiveProvider)
	}
}

// validateGigaChatConfig validates GigaChat-specific configuration.
func validateGigaChatConfig(cfg *GigaChatConfig) error {
	if cfg.ClientID == "" {
		return fmt.Errorf("client ID is required")
	}
	if cfg.ClientSecret == "" {
		return fmt.Errorf("client secret is required")
	}
	if cfg.Temperature < 0 || cfg.Temperature > 2.0 {
		return fmt.Errorf("temperature must be between 0.0 and 2.0")
	}
	if cfg.TopP < 0 || cfg.TopP > 1.0 {
		return fmt.Errorf("top_p must be between 0.0 and 1.0")
	}
	return nil
}

// validateChatGPTConfig validates ChatGPT-specific configuration.
func validateChatGPTConfig(cfg *ChatGPTConfig) error {
	if cfg.APIKey == "" {
		return fmt.Errorf("API key is required")
	}
	if cfg.Temperature < 0 || cfg.Temperature > 2.0 {
		return fmt.Errorf("temperature must be between 0.0 and 2.0")
	}
	if cfg.TopP < 0 || cfg.TopP > 1.0 {
		return fmt.Errorf("top_p must be between 0.0 and 1.0")
	}
	return nil
}

// validatePostgresConfig validates PostgreSQL configuration.
func validatePostgresConfig(cfg *PostgresConfig) error {
	if cfg.Host == "" {
		return fmt.Errorf("host is required")
	}
	if cfg.Port <= 0 || cfg.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	if cfg.User == "" {
		return fmt.Errorf("user is required")
	}
	if cfg.Password == "" {
		return fmt.Errorf("password is required")
	}
	if cfg.Database == "" {
		return fmt.Errorf("database name is required")
	}
	return nil
}

// setDefaults sets default values for optional configuration fields.
func setDefaults(cfg *BotConfig) {
	// GigaChat defaults
	if cfg.LLM.GigaChat.Model == "" {
		cfg.LLM.GigaChat.Model = "GigaChat-Max"
	}
	if cfg.LLM.GigaChat.Temperature == 0 {
		cfg.LLM.GigaChat.Temperature = 0.1
	}
	if cfg.LLM.GigaChat.TopP == 0 {
		cfg.LLM.GigaChat.TopP = 1.0
	}
	if cfg.LLM.GigaChat.MaxTokens == 0 {
		cfg.LLM.GigaChat.MaxTokens = 2048
	}
	if cfg.LLM.GigaChat.RepetitionPenalty == 0 {
		cfg.LLM.GigaChat.RepetitionPenalty = 1.0
	}

	// ChatGPT defaults
	if cfg.LLM.ChatGPT.Model == "" {
		cfg.LLM.ChatGPT.Model = "gpt-4o"
	}
	if cfg.LLM.ChatGPT.Temperature == 0 {
		cfg.LLM.ChatGPT.Temperature = 0.1
	}
	if cfg.LLM.ChatGPT.TopP == 0 {
		cfg.LLM.ChatGPT.TopP = 1.0
	}
	if cfg.LLM.ChatGPT.MaxTokens == 0 {
		cfg.LLM.ChatGPT.MaxTokens = 2048
	}

	// Storage defaults
	if cfg.Storage.History.Type == "" {
		cfg.Storage.History.Type = "postgres"
	}
	if cfg.Storage.Users.Type == "" {
		cfg.Storage.Users.Type = "postgres"
	}

	// PostgreSQL defaults
	if cfg.Storage.Postgres.SSLMode == "" {
		cfg.Storage.Postgres.SSLMode = "disable"
	}
	if cfg.Storage.Postgres.Port == 0 {
		cfg.Storage.Postgres.Port = 5432
	}
}
