package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/indienSs/tg-bot-go/internal/clients/openai"
	"github.com/indienSs/tg-bot-go/internal/config"
	"github.com/indienSs/tg-bot-go/internal/handler"
	"github.com/indienSs/tg-bot-go/internal/repository/postgres"
	"github.com/indienSs/tg-bot-go/internal/repository/redis"
	"github.com/indienSs/tg-bot-go/internal/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {

	redisDb, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		log.Fatalf("Failed to get redis_db env: %v", err)
	}

	openaiTemp64, err := strconv.ParseFloat(os.Getenv("OPENAI_TEMPERATURE"), 32)
	if err != nil {
		log.Fatalf("Failed to get openai_temperature env: %v", err)
	}
	openaiTemp := float32(openaiTemp64)

	openaiTokens, err := strconv.Atoi(os.Getenv("OPENAI_MAX_TOKENS"))
	if err != nil {
		log.Fatalf("Failed to get openai_max_tokens env: %v", err)
	}

	cfg := config.Config{
		TelegramToken: os.Getenv("TELEGRAM_TOKEN"),
		Postgres: config.PostgresConfig{
			Host:     os.Getenv("POSTGRES_HOST"),
			Port:     os.Getenv("POSTGRES_PORT"),
			User:     os.Getenv("POSTGRES_USER"),
			Password: os.Getenv("POSTGRES_PASSWORD"),
			DBName:   os.Getenv("POSTGRES_DB"),
			SSLMode:  os.Getenv("POSTGRES_SSLMODE"),
		},
		Redis: config.RedisConfig{
			Addr:     os.Getenv("REDIS_ADDR"),
			Password: os.Getenv("REDIS_PASSWORD"),
			DB:       redisDb,
		},
		OpenAI: config.OpenAIConfig{
            APIKey:      os.Getenv("OPENAI_API_KEY"),
            Model:       os.Getenv("OPENAI_MODEL"),
            Temperature: openaiTemp,
            MaxTokens:   openaiTokens,
        },
	}

	bot, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	pg, err := postgres.New(cfg.Postgres)
	if err != nil {
		log.Fatalf("Failed to connect to postgres: %v", err)
	}
	defer pg.Close()

	rdb, err := redis.New(cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to connect to redis: %v", err)
	}
	defer rdb.Close()

	aiClient := openai.New(cfg.OpenAI.APIKey, openai.OpenAIConfig{
        Model:       cfg.OpenAI.Model,
        Temperature: cfg.OpenAI.Temperature,
        MaxTokens:   cfg.OpenAI.MaxTokens,
    })

	svc := service.New(pg, rdb, aiClient)

	h := handler.New(bot, svc)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)
		<-sigint
		cancel()
	}()

	h.HandleUpdates(ctx)

	log.Println("Bot stopped")
}