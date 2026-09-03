package handler

import (
	"context"
	"cu-timepad-bot/internal/templates"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (h *Handler) About(ctx context.Context, b *bot.Bot, update *models.Update) {
	message := getMessage(update)
	templateData := h.defaultData(update)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    message.Chat.ID,
		Text:      templates.Render("about", &templateData),
		ParseMode: models.ParseModeHTML,
	})
}
