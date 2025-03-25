package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/indienSs/tg-bot-go/internal/config"
	"github.com/indienSs/tg-bot-go/internal/handler"
	"github.com/indienSs/tg-bot-go/internal/repository/postgres"
	"github.com/indienSs/tg-bot-go/internal/repository/redis"
	"github.com/indienSs/tg-bot-go/internal/service"
	"github.com/joho/godotenv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	if err := godotenv.Load(); err != nil {
        log.Fatalf("No .env file found or error loading .env: %v", err)
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
			DB:       0,
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

	svc := service.New(pg, rdb)

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