package config

// LLM represents configuration for different LLM providers.
type LLM struct {
	Provider          `yaml:"provider"`
	ValidateResponses bool `yaml:"validate_responses"`

	OpenAI    `yaml:"openai"`
	Anthropic `yaml:"anthropic"`
	DeepSeek  `yaml:"deepseek"`
	Mistral   `yaml:"mistral"`
}

// Provider defines supported LLM providers.
type Provider string

const (
	ProviderOpenAI    Provider = "openai"
	ProviderAnthropic Provider = "anthropic"
	ProviderDeepSeek  Provider = "deepseek"
	ProviderMistral   Provider = "mistral"
)

// OpenAI configuration for models.
type OpenAI struct {
	APIKey      string  `yaml:"api_key"`
	Model       string  `yaml:"model"`
	Temperature float64 `yaml:"temperature"`
	TopP        float64 `yaml:"top_p"`
	MaxTokens   int64   `yaml:"max_tokens"`
	SocksProxy  string  `yaml:"socks_proxy"`
	BaseURL     string  `yaml:"base_url"`
}

// Anthropic configuration for models.
type Anthropic struct {
	APIKey      string  `yaml:"api_key"`
	Model       string  `yaml:"model"`
	Temperature float64 `yaml:"temperature"`
	TopP        float64 `yaml:"top_p"`
	MaxTokens   int64   `yaml:"max_tokens"`
	SocksProxy  string  `yaml:"socks_proxy"`
	BaseURL     string  `yaml:"base_url"`
}

// DeepSeek configuration for models.
type DeepSeek struct {
	APIKey      string  `yaml:"api_key"`
	Model       string  `yaml:"model"`
	Temperature float64 `yaml:"temperature"`
	TopP        float64 `yaml:"top_p"`
	MaxTokens   int64   `yaml:"max_tokens"`
	SocksProxy  string  `yaml:"socks_proxy"`
	BaseURL     string  `yaml:"base_url"`
}

// Mistral configuration for models.
type Mistral struct {
	APIKey      string  `yaml:"api_key"`
	Model       string  `yaml:"model"`
	Temperature float64 `yaml:"temperature"`
	TopP        float64 `yaml:"top_p"`
	MaxTokens   int64   `yaml:"max_tokens"`
	SocksProxy  string  `yaml:"socks_proxy"`
	BaseURL     string  `yaml:"base_url"`
}
