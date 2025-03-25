package service

import (
	"context"
	"fmt"
	"time"

	"github.com/indienSs/tg-bot-go/internal/clients/openai"
	"github.com/indienSs/tg-bot-go/internal/repository/postgres"
	"github.com/indienSs/tg-bot-go/internal/repository/redis"
)

type Service struct {
	postgres *postgres.Postgres
	redis    *redis.Redis
	ai       *openai.Client
}

func New(postgres *postgres.Postgres, redis *redis.Redis, ai *openai.Client) *Service {
	return &Service{
		postgres: postgres,
		redis:    redis,
		ai:       ai,
	}
}

func (s *Service) ProcessMessage(ctx context.Context, telegramID int64, username, firstName, lastName, text string) (string, error) {
	cacheKey := fmt.Sprintf("user:%d:last_message", telegramID)
	lastMessage, err := s.redis.Get(ctx, cacheKey)
	if err == nil && lastMessage == text {
		return "", nil
	}

	if err := s.postgres.SaveUser(ctx, telegramID, username, firstName, lastName); err != nil {
		return "", fmt.Errorf("failed to save user: %w", err)
	}

	if err := s.postgres.SaveMessage(ctx, telegramID, text); err != nil {
		return "", fmt.Errorf("failed to save message: %w", err)
	}

	if err := s.redis.Set(ctx, cacheKey, text, time.Hour); err != nil {
		return "", fmt.Errorf("failed to cache message: %w", err)
	}

	response, err := s.ai.GenerateResponse(ctx, text)
	if err != nil {
		return "", fmt.Errorf("failed to generate AI response: %w", err)
	}

	return response, nil
}