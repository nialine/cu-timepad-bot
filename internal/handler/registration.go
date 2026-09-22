package handler

import (
	"context"
	"cu-timepad-bot/internal/domain"
	"cu-timepad-bot/internal/drafts"
	"cu-timepad-bot/internal/templates"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (h *Handler) handleRegistration(ctx context.Context, b *bot.Bot, update *models.Update) {
	templateData := h.defaultData(update)
	message := getMessage(update)
	if update.CallbackQuery != nil {
		message.Text = ""
	}
	status := h.svc.InRegistrationProcess(ctx, message.Chat.ID)
	var state drafts.State
	var err error
	switch status {
	case domain.StatusNone:
		fallthrough
	case domain.StatusRegistration:
		state, err = h.svc.HandleRegistrationInput(ctx, message.Chat.ID, message.Text)
	case domain.StatusTempRegistration:
		state, err = h.svc.HandleTempBookingInput(ctx, message.Chat.ID, message.Text, nil)
	}
	if state == drafts.StateEmail && err == domain.ErrDataIsInvalid {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    message.Chat.ID,
			Text:      templates.Render("invalid_email", &templateData),
			ParseMode: models.ParseModeHTML,
		})
	}

	switch state {
	case drafts.StateEmail:
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    message.Chat.ID,
			Text:      templates.Render("registration_email", &templateData),
			ParseMode: models.ParseModeHTML,
		})
	case drafts.StateFirstName:
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    message.Chat.ID,
			Text:      templates.Render("registration_firstname", &templateData),
			ParseMode: models.ParseModeHTML,
		})
	case drafts.StateLastName:
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    message.Chat.ID,
			Text:      templates.Render("registration_lastname", &templateData),
			ParseMode: models.ParseModeHTML,
		})
	case drafts.StateDone:
		if status == domain.StatusRegistration {
			kb := &models.InlineKeyboardMarkup{
				InlineKeyboard: [][]models.InlineKeyboardButton{
					{{
						Text:         templates.Render("home_button", &templateData),
						CallbackData: StartCallback,
					}},
				},
			}

			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:      message.Chat.ID,
				Text:        templates.Render("registration_done", &templateData),
				ParseMode:   models.ParseModeHTML,
				ReplyMarkup: kb,
			})
		} else {
			h.reserveSlotCallback(ctx, b, update, nil)
		}
	}
}
