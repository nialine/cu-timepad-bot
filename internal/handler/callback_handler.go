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
	case SubscribeEventsCallback:
		h.SubscribeEventsCallback(ctx, b, update, data[1:])
	case StartCallback:
		h.EditStart(ctx, b, update)
	case ChooseEventForBookingCallback:
		h.ChooseEventCallback(ctx, b, update, ChooseSlotCallback)
	case ChooseSlotCallback:
		h.chooseSlotCallback(ctx, b, update, parseChooseSlotCallback(data[1:]))
	case ReserveSlotCallback:
		h.reserveSlotCallback(ctx, b, update, parseReserveSlotCallback(data[1:]))
	case BookingDataCallback:
		h.handleRegistration(ctx, b, update)
	default:
		slog.LogAttrs(ctx,
			slog.LevelWarn,
			"Unknown callback",
			slog.String("callback", update.CallbackQuery.Data),
		)
	}
}
