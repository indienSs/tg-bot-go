package service

import (
	"context"
	"fmt"
	"time"

	"github.com/indienSs/tg-bot-go/internal/repository/postgres"
	"github.com/indienSs/tg-bot-go/internal/repository/redis"
)

type Service struct {
	postgres *postgres.Postgres
	redis    *redis.Redis
}

func New(postgres *postgres.Postgres, redis *redis.Redis) *Service {
	return &Service{
		postgres: postgres,
		redis:    redis,
	}
}

func (s *Service) ProcessMessage(ctx context.Context, telegramID int64, username, firstName, lastName, text string) error {
	cacheKey := fmt.Sprintf("user:%d:last_message", telegramID)
	lastMessage, err := s.redis.Get(ctx, cacheKey)
	if err == nil && lastMessage == text {
		return nil
	}

	if err := s.postgres.SaveUser(ctx, telegramID, username, firstName, lastName); err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}

	if err := s.postgres.SaveMessage(ctx, telegramID, text); err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}

	if err := s.redis.Set(ctx, cacheKey, text, time.Hour); err != nil {
		return fmt.Errorf("failed to cache message: %w", err)
	}

	return nil
}