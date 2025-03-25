package openai

import (
	"context"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

type Client struct {
    client *openai.Client
    config OpenAIConfig
}

type OpenAIConfig struct {
    Model       string
    Temperature float32
    MaxTokens   int
}

func New(apiKey string, cfg OpenAIConfig) *Client {
    return &Client{
        client: openai.NewClient(apiKey),
        config: cfg,
    }
}

func (c *Client) GenerateResponse(ctx context.Context, prompt string) (string, error) {
    resp, err := c.client.CreateChatCompletion(
        ctx,
        openai.ChatCompletionRequest{
            Model:       c.config.Model,
            Temperature: c.config.Temperature,
            MaxTokens:   c.config.MaxTokens,
            Messages: []openai.ChatCompletionMessage{
                {
                    Role:    openai.ChatMessageRoleUser,
                    Content: prompt,
                },
            },
        },
    )

    if err != nil {
        return "", fmt.Errorf("openai completion error: %w", err)
    }

    return resp.Choices[0].Message.Content, nil
}