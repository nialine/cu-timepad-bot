package handler

import (
	"context"
	"cu-timepad-bot/internal/config"
	"cu-timepad-bot/internal/templates"
	"fmt"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (h *Handler) ChooseEventCallback(ctx context.Context, b *bot.Bot, update *models.Update, nextCallback string) {
	templateData := h.defaultData(update)
	cfg := config.GetConfig(ctx)
	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{},
	}

	for _, ev := range cfg.Events {
		kb.InlineKeyboard = append(kb.InlineKeyboard, []models.InlineKeyboardButton{{
			Text:         ev.Name,
			CallbackData: fmt.Sprintf("%v:%v", nextCallback, ev.ID),
		}})
	}

	kb.InlineKeyboard = append(kb.InlineKeyboard, []models.InlineKeyboardButton{{
		Text:         templates.Render("back_button", &templateData),
		CallbackData: StartCallback,
	}})

	_, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      update.CallbackQuery.Message.Message.Chat.ID,
		MessageID:   update.CallbackQuery.Message.Message.ID,
		Text:        templates.Render("choose_event", &templateData),
		ReplyMarkup: kb,
	})
	if err != nil {
		h.writeError(ctx, b, update, err)
	}
}
