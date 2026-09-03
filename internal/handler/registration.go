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
	state, err := h.svc.HandleRegistrationInput(ctx, message.Chat.ID, message.Text)
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
		kb := &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{{
					Text:         templates.Render("home_button", &templateData),
					CallbackData: startCallback,
				}},
			},
		}

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      message.Chat.ID,
			Text:        templates.Render("registration_done", &templateData),
			ParseMode:   models.ParseModeHTML,
			ReplyMarkup: kb,
		})
	}
}

func (h *Handler) handleTempRegistration(ctx context.Context, b *bot.Bot, update *models.Update) {
	templateData := h.defaultData(update)
	message := getMessage(update)
	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{{
				Text:         templates.Render("home_button", &templateData),
				CallbackData: startCallback,
			}},
		},
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      message.Chat.ID,
		Text:        templates.Render("tempregistration_notimplemented", &templateData),
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: kb,
	})
}