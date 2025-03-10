package config

// LLM represents configuration for different LLM providers.
type LLM struct {
	Provider        `yaml:"provider"`
	VerifyResponses bool `yaml:"verify_responses"`

	ChatGPT   `yaml:"chatgpt"`
	Anthropic `yaml:"anthropic"`
	DeepSeek  `yaml:"deepseek"`
	Mistral   `yaml:"mistral"`
}

// Provider defines supported LLM providers.
type Provider string

const (
	ProviderChatGPT   Provider = "chatgpt"
	ProviderAnthropic Provider = "anthropic"
	ProviderDeepSeek  Provider = "deepseek"
	ProviderMistral   Provider = "mistral"
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

// Anthropic configuration for Anthropic models.
type Anthropic struct {
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

// Mistral configuration for Mistral AI models.
type Mistral struct {
	APIKey      string  `yaml:"api_key"`
	Model       string  `yaml:"model"`
	Temperature float64 `yaml:"temperature"`
	TopP        float64 `yaml:"top_p"`
	MaxTokens   int64   `yaml:"max_tokens"`
	SocksProxy  string  `yaml:"socks_proxy"`
	BaseURL     string  `yaml:"base_url"`
}
