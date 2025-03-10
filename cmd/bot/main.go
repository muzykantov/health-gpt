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

	// Initialize LLM based on configuration.
	var ai server.ChatCompleter
	switch cfg.LLM.Provider {
	case config.ProviderChatGPT:
		ai, err = llm.NewChatGPT(
			cfg.LLM.ChatGPT.APIKey,
			llm.ChatGPTWithTemperature(cfg.LLM.ChatGPT.Temperature),
			llm.ChatGPTWithModel(cfg.LLM.ChatGPT.Model),
			llm.ChatGPTWithTopP(cfg.LLM.ChatGPT.TopP),
			llm.ChatGPTWithMaxTokens(cfg.LLM.ChatGPT.MaxTokens),
			llm.ChatGPTWithSocksProxy(cfg.LLM.ChatGPT.SocksProxy),
			llm.ChatGPTWithBaseURL(cfg.LLM.ChatGPT.BaseURL),
		)
	case config.ProviderDeepSeek:
		ai, err = llm.NewDeepSeek(
			cfg.LLM.DeepSeek.APIKey,
			llm.DeepSeekWithTemperature(cfg.LLM.DeepSeek.Temperature),
			llm.DeepSeekWithModel(cfg.LLM.DeepSeek.Model),
			llm.DeepSeekWithTopP(cfg.LLM.DeepSeek.TopP),
			llm.DeepSeekWithMaxTokens(cfg.LLM.DeepSeek.MaxTokens),
			llm.DeepSeekWithSocksProxy(cfg.LLM.DeepSeek.SocksProxy),
			llm.DeepSeekWithBaseURL(cfg.LLM.DeepSeek.BaseURL),
		)
	default:
		log.Fatalf("unknown LLM provider: %s", cfg.LLM.Provider)
	}

	if err != nil {
		log.Fatalf("creating LLM client: %v", err)
	}

	// Initialize handlers.
	var (
		myGeneticsHandler = handler.MyGenetics()
		authHandler       = handler.Auth(myGeneticsHandler)
	)

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
		Handler:             authHandler,
		Completion:          ai,
		Storage:             dataStorage,
		Debug:               cfg.Telegram.Debug,
		UnsupportedResponse: unsupported,
		ErrorLog:            log.Default(),
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
}
