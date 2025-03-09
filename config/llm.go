package config

// LLM represents configuration for different LLM providers.
type LLM struct {
	Provider `yaml:"provider"`

	ChatGPT  `yaml:"chatgpt"`
	Claude   `yaml:"claude"`
	DeepSeek `yaml:"deepseek"`
}

// Provider defines supported LLM providers.
type Provider string

const (
	ProviderChatGPT  Provider = "chatgpt"
	ProviderClaudeAI Provider = "claude"
	ProviderDeepSeek Provider = "deepseek"
)

// ChatGPT configuration for OpenAI GPT models.
type ChatGPT struct {
	APIKey      string  `yaml:"api_key"`
	Model       string  `yaml:"model"`
	Temperature float64 `yaml:"temperature"`
	TopP        float64 `yaml:"top_p"`
	MaxTokens   int64   `yaml:"max_tokens"`
	SocksProxy  string  `yaml:"socks_proxy"`
	BaseURL     string  `yaml:"base_url"`
}

// DeepSeek configuration for DeepSeek models.
type DeepSeek struct {
	APIKey      string  `yaml:"api_key"`
	Model       string  `yaml:"model"`
	Temperature float64 `yaml:"temperature"`
	TopP        float64 `yaml:"top_p"`
	MaxTokens   int64   `yaml:"max_tokens"`
	SocksProxy  string  `yaml:"socks_proxy"`
	BaseURL     string  `yaml:"base_url"`
}

// Claude configuration for Anthropic models.
type Claude struct {
	APIKey      string  `yaml:"api_key"`
	Model       string  `yaml:"model"`
	Temperature float64 `yaml:"temperature"`
	MaxTokens   int     `yaml:"max_tokens"`
	// TODO: Implement ClaudeAI configuration.
}
