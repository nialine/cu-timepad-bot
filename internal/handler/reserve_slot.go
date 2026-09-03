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

type reserveSlotData struct {
	eventid int64
	page    int
	slotid  int64
}

func parseReserveSlotCallback(callbackData []string) *reserveSlotData {
	if len(callbackData) >= 3 {
		var data reserveSlotData
		data.eventid, _ = strconv.ParseInt(callbackData[0], 10, 64)
		data.page, _ = strconv.Atoi(callbackData[1])
		data.slotid, _ = strconv.ParseInt(callbackData[2], 10, 64)
		return &data
	}
	return nil
}

func (h *Handler) reserveSlotCallback(ctx context.Context, b *bot.Bot, update *models.Update, data *reserveSlotData) {
	templateData := h.defaultData(update)

	if data == nil {
		h.writeError(ctx, b, update, domain.ErrDataNotSufficient)
		return
	}

	kb := models.InlineKeyboardMarkup{InlineKeyboard: [][]models.InlineKeyboardButton{{{
		Text:         templates.Render("back_button", &templateData),
		CallbackData: fmt.Sprintf("%v:%v:%v", chooseSlotCallback, data.eventid, data.page),
	}, {
		Text:         templates.Render("home_button", &templateData),
		CallbackData: startCallback,
	}}}}

	message := getMessage(update)
	userid := message.Chat.ID
	if !h.svc.HasRegistrationData(ctx, userid) {
		slog.LogAttrs(ctx,
			slog.LevelError,
			"TempRegistration: Not implemented",
		)

	}
	slot, err := h.svc.ReserveSlot(ctx, userid, data.eventid, data.slotid, nil)
	if err != nil {
		b.EditMessageText(ctx, &bot.EditMessageTextParams{
			ChatID:      message.Chat.ID,
			MessageID:   message.ID,
			Text:        templates.Render("error_reserved_slot", &templateData),
			ReplyMarkup: kb,
			ParseMode:   models.ParseModeHTML,
		})
		slog.LogAttrs(ctx,
			slog.LevelError,
			"Error reserving slot",
			slog.Any("error", err),
		)
		return
	}
	templateData["slot"] = slot

	b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      message.Chat.ID,
		MessageID:   message.ID,
		Text:        templates.Render("reserved_slot", &templateData),
		ReplyMarkup: kb,
		ParseMode:   models.ParseModeHTML,
	})
}
