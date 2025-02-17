package config

// LLM represents configuration for different LLM providers.
type LLM struct {
	Provider `yaml:"provider" validate:"required,oneof=chatgpt claude deepseek"`

	ChatGPT  `yaml:"chatgpt" validate:"required_if=Provider chatgpt"`
	Claude   `yaml:"claude" validate:"required_if=Provider claude"`
	DeepSeek `yaml:"deepseek" validate:"required_if=Provider deepseek"`
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
	APIKey      string  `yaml:"api_key" validate:"required"`
	Model       string  `yaml:"model" validate:"omitempty"`
	Temperature float64 `yaml:"temperature" validate:"omitempty,min=0,max=2"`
	TopP        float64 `yaml:"top_p" validate:"omitempty,min=0,max=1"`
	MaxTokens   int     `yaml:"max_tokens" validate:"omitempty,min=1,max=32000"`
	SocksProxy  string  `yaml:"socks_proxy" validate:"omitempty,url"`
}

// Claude configuration for Anthropic models.
// TODO: Implement ClaudeAI configuration.
type Claude struct {
	APIKey      string  `yaml:"api_key" validate:"required"`
	Model       string  `yaml:"model" validate:"omitempty"`
	Temperature float64 `yaml:"temperature" validate:"omitempty,min=0,max=1"`
	MaxTokens   int     `yaml:"max_tokens" validate:"omitempty,min=1,max=200000"`
}

// DeepSeek configuration for DeepSeek models.
// TODO: Implement DeepSeek configuration.
type DeepSeek struct {
	APIKey      string  `yaml:"api_key" validate:"required"`
	Model       string  `yaml:"model" validate:"omitempty"`
	Temperature float64 `yaml:"temperature" validate:"omitempty,min=0,max=1"`
	MaxTokens   int     `yaml:"max_tokens" validate:"omitempty,min=1,max=32000"`
}
