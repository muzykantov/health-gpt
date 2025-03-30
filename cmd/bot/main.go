package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"

	"github.com/muzykantov/health-gpt/chat"
	"github.com/muzykantov/health-gpt/chat/storage"
	"github.com/muzykantov/health-gpt/config"
	"github.com/muzykantov/health-gpt/handler"
	"github.com/muzykantov/health-gpt/llm"
	"github.com/muzykantov/health-gpt/metrics"
	"github.com/muzykantov/health-gpt/server"
	"github.com/muzykantov/health-gpt/server/telegram"
)

const MsgUnsupportedType = "❌ Тип сообщения не поддерживается."

func main() {
	// Parse command line flags.
	configPath := flag.String("config", "config.yaml", "path to configuration file")
	flag.Parse()

	// Load configuration.
	cfg, err := config.FromFile(*configPath)
	if err != nil {
		log.Fatalf("loading configuration: %v", err)
	}

	logger := log.Default()

	// Start metrics server if enabled.
	var metricsServer *metrics.Server
	if cfg.Metrics.Enabled {
		metricsAddress := cfg.Metrics.Address
		if metricsAddress == "" {
			metricsAddress = ":9090" // default value
		}

		metricsServer = metrics.NewServer(metricsAddress, logger)
		go func() {
			logger.Printf("Starting metrics server on %s", metricsAddress)
			if err := metricsServer.Start(); err != nil {
				logger.Printf("Metrics server error: %v", err)
			}
		}()
	}

	// Initialize LLM based on configuration.
	var (
		ai    server.ChatCompleter
		model string
	)
	switch cfg.LLM.Provider {
	case config.ProviderOpenAI:
		model = cfg.LLM.OpenAI.Model
		ai, err = llm.NewOpenAI(
			cfg.LLM.OpenAI.APIKey,
			llm.OpenAIWithTemperature(cfg.LLM.OpenAI.Temperature),
			llm.OpenAIWithModel(cfg.LLM.OpenAI.Model),
			llm.OpenAIWithTopP(cfg.LLM.OpenAI.TopP),
			llm.OpenAIWithMaxTokens(cfg.LLM.OpenAI.MaxTokens),
			llm.OpenAIWithSocksProxy(cfg.LLM.OpenAI.SocksProxy),
			llm.OpenAIWithBaseURL(cfg.LLM.OpenAI.BaseURL),
		)
	case config.ProviderAnthropic:
		model = cfg.LLM.Anthropic.Model
		ai, err = llm.NewAnthropic(
			cfg.LLM.Anthropic.APIKey,
			llm.AnthropicWithTemperature(cfg.LLM.Anthropic.Temperature),
			llm.AnthropicWithModel(cfg.LLM.Anthropic.Model),
			llm.AnthropicWithTopP(cfg.LLM.Anthropic.TopP),
			llm.AnthropicWithMaxTokens(cfg.LLM.Anthropic.MaxTokens),
			llm.AnthropicWithSocksProxy(cfg.LLM.Anthropic.SocksProxy),
			llm.AnthropicWithBaseURL(cfg.LLM.Anthropic.BaseURL),
		)
	case config.ProviderDeepSeek:
		model = cfg.LLM.DeepSeek.Model
		ai, err = llm.NewDeepSeek(
			cfg.LLM.DeepSeek.APIKey,
			llm.DeepSeekWithTemperature(cfg.LLM.DeepSeek.Temperature),
			llm.DeepSeekWithModel(cfg.LLM.DeepSeek.Model),
			llm.DeepSeekWithTopP(cfg.LLM.DeepSeek.TopP),
			llm.DeepSeekWithMaxTokens(cfg.LLM.DeepSeek.MaxTokens),
			llm.DeepSeekWithSocksProxy(cfg.LLM.DeepSeek.SocksProxy),
			llm.DeepSeekWithBaseURL(cfg.LLM.DeepSeek.BaseURL),
		)
	case config.ProviderMistral:
		model = cfg.LLM.Mistral.Model
		ai, err = llm.NewMistral(
			cfg.LLM.Mistral.APIKey,
			llm.MistralWithTemperature(cfg.LLM.Mistral.Temperature),
			llm.MistralWithModel(cfg.LLM.Mistral.Model),
			llm.MistralWithTopP(cfg.LLM.Mistral.TopP),
			llm.MistralWithMaxTokens(cfg.LLM.Mistral.MaxTokens),
			llm.MistralWithSocksProxy(cfg.LLM.Mistral.SocksProxy),
			llm.MistralWithBaseURL(cfg.LLM.Mistral.BaseURL),
		)
	default:
		log.Fatalf("unknown LLM provider: %s", cfg.LLM.Provider)
	}

	if err != nil {
		log.Fatalf("creating LLM client: %v", err)
	}

	if cfg.LLM.ValidateResponses {
		ai = llm.NewValidator(
			ai,
			ai,
			string(cfg.LLM.Provider), model,
			string(cfg.LLM.Provider), model,
			0, // default value
			cfg.Telegram.Debug,
			logger,
		)
	}

	var dataStorage server.DataStorage
	switch cfg.Storage.Type {
	case config.TypeFS:
		dataStorage, err = storage.NewFS(cfg.Storage.FS.Dir)
		if err != nil {
			log.Fatalf("creating file storage: %v", err)
		}

	case config.TypeBolt:
		boltStorage, err := storage.NewBolt(cfg.Storage.Bolt.Path)
		if err != nil {
			log.Fatalf("creating bolt db storage: %v", err)
		}
		dataStorage = boltStorage
		defer boltStorage.Close()

	default:
		log.Fatalf("unknown storage type: %s", cfg.Storage.Type)
	}

	unsupported := func() chat.Message {
		return chat.NewMessage(chat.RoleAssistant, MsgUnsupportedType)
	}

	// Create and configure the server.
	srv := &telegram.Server{
		Token:               cfg.Telegram.Token,
		Handler:             handler.Start(),
		Completion:          ai,
		Storage:             dataStorage,
		Debug:               cfg.Telegram.Debug,
		UnsupportedResponse: unsupported,
		Log:                 logger,
	}

	// Setup context with signal handling.
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	// Start the server.
	if err := srv.ListenAndServe(ctx); err != nil {
		log.Fatalf("starting server: %v", err)
	}

	// Gracefully shutdown metrics server if it was started
	if metricsServer != nil {
		if err := metricsServer.Shutdown(context.Background()); err != nil {
			logger.Printf("Error shutting down metrics server: %v", err)
		}
	}
}
