package handler

import (
	"context"
	"cu-timepad-bot/internal/domain"
	"log/slog"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Store interface {
	AddUser(ctx context.Context, userid int64) error
	GetUser(ctx context.Context, userid int64) (*domain.User, error)
	FindUsersWithEvent(ctx context.Context, event domain.EventID) []domain.User
	AddUserSubscribedEvent(ctx context.Context, userid int64, event domain.EventID) error
	RemoveUserSubscribedEvent(ctx context.Context, userid int64, event domain.EventID) error
}

type Handler struct {
	st Store
}

func New(st Store) Handler {
	return Handler{
		st,
	}
}

func (h *Handler) writeError(ctx context.Context, b *bot.Bot, update *models.Update, err error) {
	if err != nil {
		slog.LogAttrs(ctx,
			slog.LevelError,
			"Coundn't send message",
			slog.Any("error", err),
			slog.Int64("chatid", update.Message.Chat.ID),
		)
		data := h.defaultData(update)
		text, _ := renderTemplate("error", data)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    update.Message.Chat.ID,
			Text:      text,
			ParseMode: models.ParseModeHTML,
		})
	}
}

func (h *Handler) defaultData(update *models.Update) map[string]any {
	return map[string]any{
		"chatid":        update.Message.Chat.ID,
		"message_text":  update.Message.Text,
		"first_name_tg": update.Message.Chat.FirstName,
		"last_name_tg":  update.Message.Chat.LastName,
		"time":          time.Now(),
	}
}

func (h *Handler) Start(ctx context.Context, b *bot.Bot, update *models.Update) {
	data := h.defaultData(update)
	text, err := renderTemplate("start", data)
	h.writeError(ctx, b, update, err)

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		ReplyParameters: &models.ReplyParameters{
			MessageID: update.Message.ID,
		},
		Text:      text,
		ParseMode: models.ParseModeHTML,
	})
	h.writeError(ctx, b, update, err)
}
