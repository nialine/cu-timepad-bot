package handler

import (
	"context"
	"log/slog"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (h *Handler) CallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       false,
	})

	data := strings.Split(update.CallbackQuery.Data, ":")
	switch data[0] {
	case subscribeEventsCallback:
		h.SubscribeEventsCallback(ctx, b, update, data[1:])
	case startCallback:
		h.EditStart(ctx, b, update)
	case chooseEventForBookingCallback:
		h.ChooseEventCallback(ctx, b, update, chooseSlotCallback)
	case chooseSlotCallback:
		h.chooseSlotCallback(ctx, b, update, parseChooseSlotCallback(data[1:]))
	case reserveSlotCallback:
		h.reserveSlotCallback(ctx, b, update, parseReserveSlotCallback(data[1:]))
	case bookingDataCallback:
		h.handleRegistration(ctx, b, update)
	default:
		slog.LogAttrs(ctx,
			slog.LevelWarn,
			"Unknown callback",
			slog.String("callback", update.CallbackQuery.Data),
		)
	}
}
