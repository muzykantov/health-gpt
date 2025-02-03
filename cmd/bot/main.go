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
	configPath := flag.String("config", "config.json", "path to configuration file")
	flag.Parse()

	// Загружаем конфигурацию.
	cfg, err := config.LoadBotConfigFromFile(*configPath)
	if err != nil {
		fmt.Printf("Ошибка загрузки конфигурации: %v\n", err)
		os.Exit(1)
	}

	// Инициализируем LLM в зависимости от конфигурации.
	var ai server.ChatCompleter
	switch cfg.LLM.ActiveProvider {
	case "gigachat":
		ai, err = llm.NewGigaChat(
			cfg.LLM.GigaChat.ClientID,
			cfg.LLM.GigaChat.ClientSecret,
			llm.GigaChatWithTemperature(cfg.LLM.GigaChat.Temperature),
			llm.GigaChatWithModel(cfg.LLM.GigaChat.Model),
			llm.GigaChatWithTopP(cfg.LLM.GigaChat.TopP),
			llm.GigaChatWithMaxTokens(cfg.LLM.GigaChat.MaxTokens),
			llm.GigaChatWithRepetitionPenalty(cfg.LLM.GigaChat.RepetitionPenalty),
		)
	case "chatgpt":
		ai, err = llm.NewChatGPT(
			cfg.LLM.ChatGPT.APIKey,
			llm.ChatGPTWithTemperature(cfg.LLM.ChatGPT.Temperature),
			llm.ChatGPTWithModel(cfg.LLM.ChatGPT.Model),
			llm.ChatGPTWithTopP(cfg.LLM.ChatGPT.TopP),
			llm.ChatGPTWithMaxTokens(cfg.LLM.ChatGPT.MaxTokens),
			llm.ChatGPTWithSocksProxy(cfg.LLM.ChatGPT.SocksProxy),
		)
	default:
		fmt.Printf("Неизвестный провайдер LLM: %s\n", cfg.LLM.ActiveProvider)
		os.Exit(1)
	}
	if err != nil {
		fmt.Printf("Ошибка создания клиента LLM: %v\n", err)
		os.Exit(1)
	}

	// Инициализируем обработчики.
	var (
		myGeneticsHandler = handler.MyGenetics()
		authHandler       = handler.Auth(myGeneticsHandler)
	)

	// Инициализируем хранилище истории.
	var historyStorage server.ChatHistoryReadWriter
	switch cfg.Storage.History.Type {
	case "file":
		historyStorage, err = history.NewFileStorage(cfg.Storage.History.Path)
	case "memory":
		historyStorage = history.NewInMemory()
	default:
		fmt.Printf("Неизвестный тип хранилища истории: %s\n", cfg.Storage.History.Type)
		os.Exit(1)
	}
	if err != nil {
		fmt.Printf("Ошибка создания хранилища истории: %v\n", err)
		os.Exit(1)
	}

	// Инициализируем хранилище пользователей.
	var userStorage server.UserStorage
	switch cfg.Storage.Users.Type {
	case "file":
		userStorage, err = user.NewFileStorage(cfg.Storage.Users.Path)
	case "memory":
		userStorage = user.NewInMemory()
	default:
		fmt.Printf("Неизвестный тип хранилища пользователей: %s\n", cfg.Storage.Users.Type)
		os.Exit(1)
	}
	if err != nil {
		fmt.Printf("Ошибка создания хранилища пользователей: %v\n", err)
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
		fmt.Printf("Ошибка сервера: %v\n", err)
		os.Exit(1)
	}
}
