package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/muzykantov/health-gpt/chat"
	"github.com/muzykantov/health-gpt/chat/history"
	"github.com/muzykantov/health-gpt/chat/user"
	"github.com/muzykantov/health-gpt/config"
	"github.com/muzykantov/health-gpt/handler"
	"github.com/muzykantov/health-gpt/llm"
	"github.com/muzykantov/health-gpt/server"
	"github.com/muzykantov/health-gpt/server/telegram"
)

const MsgUnsupportedType = "❌ Тип сообщения не поддерживается."

func main() {
	// Парсим флаги командной строки.
	configPath := flag.String("config", "config.yaml", "path to configuration file")
	flag.Parse()

	// Загружаем конфигурацию.
	cfg, err := config.FromFile(*configPath)
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Инициализируем LLM в зависимости от конфигурации.
	var ai server.ChatCompleter
	switch cfg.LLM.Provider {
	case config.ProviderChatGPT:
		ai, err = llm.NewChatGPT(
			cfg.LLM.ChatGPT.APIKey,
			llm.ChatGPTWithTemperature(float32(cfg.LLM.ChatGPT.Temperature)),
			llm.ChatGPTWithModel(cfg.LLM.ChatGPT.Model),
			llm.ChatGPTWithTopP(float32(cfg.LLM.ChatGPT.TopP)),
			llm.ChatGPTWithMaxTokens(cfg.LLM.ChatGPT.MaxTokens),
			llm.ChatGPTWithSocksProxy(cfg.LLM.ChatGPT.SocksProxy),
		)
	default:
		fmt.Printf("Unknown LLM provider: %s\n", cfg.LLM.Provider)
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("Error creating LLM client: %v\n", err)
		os.Exit(1)
	}

	// Инициализируем обработчики.
	var (
		myGeneticsHandler = handler.MyGenetics()
		authHandler       = handler.Auth(myGeneticsHandler)
	)

	var (
		historyStorage server.ChatHistoryReadWriter
		userStorage    server.UserStorage
	)
	switch cfg.Storage.Type {
	case config.TypeFS:
		historyStorage, err = history.NewFileStorage(cfg.Storage.Filesystem.Path)
		if err != nil {
			fmt.Printf("Error creating history storage: %v\n", err)
			os.Exit(1)
		}

		userStorage, err = user.NewFileStorage(cfg.Storage.Filesystem.Path)
		if err != nil {
			fmt.Printf("Error creating user storage: %v\n", err)
			os.Exit(1)
		}

	case config.TypeInMemory:
		historyStorage = history.NewInMemory()
		userStorage = user.NewInMemory()

	default:
		fmt.Printf("Unknown storage type: %s\n", cfg.Storage.Type)
		os.Exit(1)
	}

	unsupported := func() chat.Message {
		return chat.NewMessage(chat.RoleAssistant, MsgUnsupportedType)
	}

	// Создаем и конфигурируем сервер.
	srv := &telegram.Server{
		Token:               cfg.Telegram.Token,
		Handler:             authHandler,
		Completion:          ai,
		History:             historyStorage,
		User:                userStorage,
		Debug:               cfg.Telegram.Debug,
		UnsupportedResponse: unsupported,
		ErrorLog:            log.Default(),
	}

	// Настраиваем контекст с обработкой сигналов.
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	// Запускаем сервер.
	if err := srv.ListenAndServe(ctx); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
		os.Exit(1)
	}
}
