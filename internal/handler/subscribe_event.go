package handler

import (
	"context"
	"cu-timepad-bot/internal/config"
	"cu-timepad-bot/internal/templates"
	"fmt"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (h *Handler) SubscribeEventsCallback(ctx context.Context, b *bot.Bot, update *models.Update, callbackData []string) {
	templateData := h.defaultData(update)
	cfg := config.GetConfig(ctx)
	userid := update.CallbackQuery.Message.Message.Chat.ID

	if len(callbackData) > 0 {
		if err := h.svc.ProcessEventCallback(ctx, userid, callbackData); err != nil {
			h.writeError(ctx, b, update, err)
			return
		}
	}

	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{},
	}

	for _, ev := range cfg.Events {
		text, err := h.svc.GenSubscribeEventButton(ctx, userid, ev)
		if err != nil {
			h.writeError(ctx, b, update, err)
			return
		}
		kb.InlineKeyboard = append(kb.InlineKeyboard, []models.InlineKeyboardButton{{
			Text:         text,
			CallbackData: fmt.Sprintf("%v:%v", subscribeEventsCallback, ev.ID),
		}})
	}

	kb.InlineKeyboard = append(kb.InlineKeyboard, []models.InlineKeyboardButton{{
		Text:         templates.Render("back_button", &templateData),
		CallbackData: startCallback,
	}})

	_, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      userid,
		MessageID:   update.CallbackQuery.Message.Message.ID,
		Text:        templates.Render("show_events", &templateData),
		ReplyMarkup: kb,
	})
	if err != nil {
		h.writeError(ctx, b, update, err)
	}
}
