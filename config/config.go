package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// BotConfig представляет основную конфигурацию приложения.
type BotConfig struct {
	// LLM - обязательная секция, должен быть указан ActiveProvider.
	LLM LLMConfig
	// Telegram - обязательная секция, требуется токен.
	Telegram TelegramConfig
	// Storage - обязательная секция.
	Storage StorageConfig
}

// LLMConfig описывает настройки языковой модели.
type LLMConfig struct {
	// ActiveProvider - обязательное поле, указывает провайдера: "gigachat" или "chatgpt".
	ActiveProvider string
	// GigaChat используется при ActiveProvider == "gigachat".
	GigaChat GigaChatConfig
	// ChatGPT используется при ActiveProvider == "chatgpt".
	ChatGPT ChatGPTConfig
}

// GigaChatConfig содержит настройки для GigaChat.
type GigaChatConfig struct {
	// ClientID - обязательное поле для аутентификации.
	ClientID string
	// ClientSecret - обязательное поле для аутентификации.
	ClientSecret string
	// Model - опционально, по умолчанию "GigaChat-Max".
	Model string
	// Temperature - опционально, по умолчанию 0.1 (диапазон 0.0-2.0).
	Temperature float64
	// TopP - опционально, по умолчанию 1.0 (диапазон 0.0-1.0).
	TopP float64
	// MaxTokens - опционально, по умолчанию 2048.
	MaxTokens int64
	// RepetitionPenalty - опционально, по умолчанию 1.0.
	RepetitionPenalty float64
}

// ChatGPTConfig содержит настройки для ChatGPT.
type ChatGPTConfig struct {
	// APIKey - обязательное поле для аутентификации.
	APIKey string
	// Model - опционально, по умолчанию "gpt-4o".
	Model string
	// Temperature - опционально, по умолчанию 0.1 (диапазон 0.0-2.0).
	Temperature float32
	// TopP - опционально, по умолчанию 1.0 (диапазон 0.0-1.0).
	TopP float32
	// MaxTokens - опционально, по умолчанию 2048.
	MaxTokens int
	// SocksProxy - опционально, по умолчанию пустая строка.
	SocksProxy string
}

// TelegramConfig описывает настройки Telegram бота.
type TelegramConfig struct {
	// Token - обязательное поле, токен доступа к Telegram API.
	Token string
	// Debug - опционально, по умолчанию true.
	Debug bool
}

// StorageConfig описывает настройки хранилища.
type StorageConfig struct {
	// История чатов.
	History HistoryStorageConfig
	// Данные пользователей.
	Users UserStorageConfig
}

// HistoryStorageConfig описывает настройки хранилища истории.
type HistoryStorageConfig struct {
	// Type - опционально, по умолчанию "memory". Доступные значения: "file", "memory".
	Type string
	// Path - обязательно только при Type == "file", по умолчанию "./data/history".
	Path string
}

// UserStorageConfig описывает настройки хранилища пользователей.
type UserStorageConfig struct {
	// Type - опционально, по умолчанию "file". Доступные значения: "file", "memory".
	Type string
	// Path - обязательно только при Type == "file", по умолчанию "./data/users".
	Path string
}

// LoadConfig читает конфигурацию из io.Reader и затем применяет переменные окружения
func LoadConfig(r io.Reader) (*BotConfig, error) {
	var cfg BotConfig
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("ошибка декодирования JSON: %w", err)
	}

	// Применяем переменные окружения поверх значений из JSON
	applyEnvironmentVariables(&cfg)

	if err := validateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("ошибка валидации конфига: %w", err)
	}

	setDefaults(&cfg)
	return &cfg, nil
}

// LoadBotConfigFromFile читает конфигурацию из файла
func LoadBotConfigFromFile(filename string) (*BotConfig, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия файла: %w", err)
	}
	defer file.Close()

	return LoadConfig(file)
}

// applyEnvironmentVariables применяет значения из переменных окружения
func applyEnvironmentVariables(cfg *BotConfig) {
	// LLM Configuration
	if provider := os.Getenv("LLM_ACTIVE_PROVIDER"); provider != "" {
		cfg.LLM.ActiveProvider = provider
	}

	// GigaChat Configuration
	if clientID := os.Getenv("GIGACHAT_CLIENT_ID"); clientID != "" {
		cfg.LLM.GigaChat.ClientID = clientID
	}
	if clientSecret := os.Getenv("GIGACHAT_CLIENT_SECRET"); clientSecret != "" {
		cfg.LLM.GigaChat.ClientSecret = clientSecret
	}

	// ChatGPT Configuration
	if apiKey := os.Getenv("CHATGPT_API_KEY"); apiKey != "" {
		cfg.LLM.ChatGPT.APIKey = apiKey
	}

	// Telegram Configuration
	if token := os.Getenv("TELEGRAM_BOT_TOKEN"); token != "" {
		cfg.Telegram.Token = token
	}
}

// validateConfig проверяет наличие обязательных полей
func validateConfig(cfg *BotConfig) error {
	// Проверка LLM секции
	if cfg.LLM.ActiveProvider == "" {
		return fmt.Errorf("не указан ActiveProvider в секции LLM")
	}

	switch cfg.LLM.ActiveProvider {
	case "gigachat":
		if cfg.LLM.GigaChat.ClientID == "" {
			return fmt.Errorf("не указан ClientID для GigaChat")
		}
		if cfg.LLM.GigaChat.ClientSecret == "" {
			return fmt.Errorf("не указан ClientSecret для GigaChat")
		}
	case "chatgpt":
		if cfg.LLM.ChatGPT.APIKey == "" {
			return fmt.Errorf("не указан APIKey для ChatGPT")
		}
	default:
		return fmt.Errorf("неизвестный провайдер LLM: %s", cfg.LLM.ActiveProvider)
	}

	// Проверка Telegram секции
	if cfg.Telegram.Token == "" {
		return fmt.Errorf("не указан Token в секции Telegram")
	}

	// Проверка путей в Storage при использовании file
	if cfg.Storage.History.Type == "file" && cfg.Storage.History.Path == "" {
		return fmt.Errorf("не указан Path для file storage в секции History")
	}
	if cfg.Storage.Users.Type == "file" && cfg.Storage.Users.Path == "" {
		return fmt.Errorf("не указан Path для file storage в секции Users")
	}

	return nil
}

// setDefaults устанавливает значения по умолчанию для опциональных полей
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
		cfg.LLM.ChatGPT.Model = "gpt-4"
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
		cfg.Storage.History.Type = "memory"
	}
	if cfg.Storage.History.Type == "file" && cfg.Storage.History.Path == "" {
		cfg.Storage.History.Path = "./data/history"
	}
	if cfg.Storage.Users.Type == "" {
		cfg.Storage.Users.Type = "file"
	}
	if cfg.Storage.Users.Type == "file" && cfg.Storage.Users.Path == "" {
		cfg.Storage.Users.Path = "./data/users"
	}
}
