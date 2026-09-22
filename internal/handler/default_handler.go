package handler

import (
	"context"
	"cu-timepad-bot/internal/domain"
	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (h *Handler) DefaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if h.svc.InRegistrationProcess(ctx, update.Message.Chat.ID) != domain.StatusNone {
		h.handleRegistration(ctx, b, update)
		return
	}
	slog.LogAttrs(ctx,
		slog.LevelWarn,
		"Invalid message",
		slog.Int64("userid", update.Message.Chat.ID),
	)
}
