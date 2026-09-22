package handler

import (
	"context"
	"cu-timepad-bot/internal/domain"
	"cu-timepad-bot/internal/templates"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func parseReserveSlotCallback(callbackData []string) *domain.ChooseSlotData {
	if len(callbackData) >= 3 {
		var data domain.ChooseSlotData
		data.EventID, _ = strconv.ParseInt(callbackData[0], 10, 64)
		data.Page, _ = strconv.Atoi(callbackData[1])
		data.SlotID, _ = strconv.ParseInt(callbackData[2], 10, 64)
		return &data
	}
	return nil
}

func (h *Handler) editOrSendMessage(ctx context.Context, b *bot.Bot, message *models.Message, text string, kb models.InlineKeyboardMarkup) {
	if !message.From.IsBot {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      message.Chat.ID,
			Text:        text,
			ReplyMarkup: kb,
			ParseMode:   models.ParseModeHTML,
		})
	} else {
		b.EditMessageText(ctx, &bot.EditMessageTextParams{
			ChatID:      message.Chat.ID,
			MessageID:   message.ID,
			Text:        text,
			ReplyMarkup: kb,
			ParseMode:   models.ParseModeHTML,
		})
	}
}

func (h *Handler) reserveSlotCallback(ctx context.Context, b *bot.Bot, update *models.Update, data *domain.ChooseSlotData) {
	templateData := h.defaultData(update)
	message := getMessage(update)
	userid := message.Chat.ID

	var booking_data *domain.BookingUserData
	if data == nil {
		draft, err := h.svc.GetTempRegistration(ctx, userid)
		if err != nil {
			h.writeError(ctx, b, update, domain.ErrDataNotSufficient)
		}
		data = &draft.ChooseSlotData
		booking_data = &draft.BookingUserData
	}

	if !h.svc.HasRegistrationData(ctx, userid) && booking_data == nil {
		h.svc.HandleTempBookingInput(ctx, userid, "", data)
		h.handleRegistration(ctx, b, update)
		return
	}

	kb := models.InlineKeyboardMarkup{InlineKeyboard: [][]models.InlineKeyboardButton{{{
		Text:         templates.Render("back_button", &templateData),
		CallbackData: fmt.Sprintf("%v:%v:%v", ChooseSlotCallback, data.EventID, data.Page),
	}, {
		Text:         templates.Render("home_button", &templateData),
		CallbackData: StartCallback,
	}}}}

	slot, err := h.svc.ReserveSlot(ctx, userid, data.EventID, data.SlotID, booking_data)
	if err != nil {
		h.editOrSendMessage(
			ctx, b, message,
			templates.Render("error_reserved_slot", &templateData),
			kb,
		)
		slog.LogAttrs(ctx,
			slog.LevelError,
			"Error reserving slot",
			slog.Any("error", err),
		)
		return
	}
	templateData["slot"] = slot
	h.editOrSendMessage(
		ctx, b, message,
		templates.Render("reserved_slot", &templateData),
		kb,
	)
}
