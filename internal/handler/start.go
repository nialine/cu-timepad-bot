package handler

import (
	"context"
	"cu-timepad-bot/internal/templates"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func makeKeyboardStart(templateData *map[string]any) *models.InlineKeyboardMarkup {
	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{
					Text:         templates.Render("events_button", templateData),
					CallbackData: SubscribeEventsCallback,
				},
			},
			{
				{
					Text:         templates.Render("reserve_slot_button", templateData),
					CallbackData: ChooseEventForBookingCallback,
				},
			},
			{
				{
					Text:         templates.Render("booking_data", templateData),
					CallbackData: BookingDataCallback,
				},
			},
		},
	}
	return kb
}

func (h *Handler) EditStart(ctx context.Context, b *bot.Bot, update *models.Update) {
	message := getMessage(update)

	templateData := h.defaultData(update)
	templateData["has_booking_data"] = h.svc.HasRegistrationData(ctx, message.Chat.ID)

	kb := makeKeyboardStart(&templateData)

	b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      message.Chat.ID,
		MessageID:   message.ID,
		Text:        templates.Render("start", &templateData),
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: kb,
	})
}

func (h *Handler) Start(ctx context.Context, b *bot.Bot, update *models.Update) {
	userid := getMessage(update).Chat.ID
	if _, err := h.svc.GetUser(ctx, userid); err != nil {
		err := h.svc.AddUser(ctx, userid)
		if err != nil {
			h.writeError(ctx, b, update, err)
			return
		}
	}

	templateData := h.defaultData(update)
	templateData["has_booking_data"] = h.svc.HasRegistrationData(ctx, userid)

	kb := makeKeyboardStart(&templateData)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      userid,
		Text:        templates.Render("start", &templateData),
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: kb,
	})
}
