package handler

import (
	"context"
	"log"

	"github.com/indienSs/tg-bot-go/internal/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Handler struct {
	bot     *tgbotapi.BotAPI
	service *service.Service
}

func New(bot *tgbotapi.BotAPI, service *service.Service) *Handler {
	return &Handler{
		bot:     bot,
		service: service,
	}
}

func (h *Handler) HandleUpdates(ctx context.Context) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := h.bot.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			return
		case update := <-updates:
			if update.Message == nil {
				continue
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
			
			// Обрабатываем сообщение
			err := h.service.ProcessMessage(
				ctx,
				update.Message.From.ID,
				update.Message.From.UserName,
				update.Message.From.FirstName,
				update.Message.From.LastName,
				update.Message.Text,
			)

			if err != nil {
				log.Printf("Failed to process message: %v", err)
				msg.Text = "Произошла ошибка при обработке сообщения"
			} else {
				msg.Text = "Сообщение получено и сохранено!"
			}

			if _, err := h.bot.Send(msg); err != nil {
				log.Printf("Failed to send message: %v", err)
			}
		}
	}
}