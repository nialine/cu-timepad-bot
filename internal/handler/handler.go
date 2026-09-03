package handler

import (
	"context"
	"cu-timepad-bot/internal/adapters/timepad"
	"cu-timepad-bot/internal/domain"
	"cu-timepad-bot/internal/drafts"
	"cu-timepad-bot/internal/templates"
	"log/slog"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const slotsOnPage = 6

type callbacks string

const (
	subscribeEventsCallback       = "events"
	startCallback                 = "start"
	chooseEventForBookingCallback = "chooseforbooking"
	chooseSlotCallback            = "selectslot"
	reserveSlotCallback           = "reserveslot"
	bookingDataCallback           = "bookingdata"
	tempBookingDataCallback       = "tempbookingdata"
)

type service interface {
	AddUser(ctx context.Context, userid int64) error
	GetUser(ctx context.Context, userid int64) (*domain.User, error)
	ProcessEventCallback(ctx context.Context, userid int64, callbackData []string) error
	GenSubscribeEventButton(ctx context.Context, userid int64, ev *domain.Event) (string, error)
	GetAvaliableSlots(ctx context.Context, ev *domain.Event) ([]timepad.RecurringEvent, error)
	HasRegistrationData(ctx context.Context, userid int64) bool
	ReserveSlot(ctx context.Context, userid int64, eventid int64, slotid int64, userdata *domain.BookingUserData) (*timepad.RecurringEvent, error)
	InRegistrationProcess(ctx context.Context, userid int64) bool
	HandleRegistrationInput(ctx context.Context, userid int64, text string) (drafts.State, error)
}

type Handler struct {
	svc service

	NewSlots chan domain.NewSlots
}

func New(st service) Handler {
	return Handler{
		svc:      st,
		NewSlots: make(chan domain.NewSlots, 4),
	}
}

func getMessage(update *models.Update) *models.Message {
	message := update.Message
	if message == nil && update.CallbackQuery != nil && update.CallbackQuery.Message.Message != nil {
		message = update.CallbackQuery.Message.Message
	}
	return message
}

func (h *Handler) writeError(ctx context.Context, b *bot.Bot, update *models.Update, err error) {
	message := getMessage(update)

	slog.LogAttrs(ctx,
		slog.LevelError,
		"Coundn't send message",
		slog.Any("error", err),
		func() slog.Attr {
			if message != nil {
				return slog.Int64("chatid", message.Chat.ID)
			}
			return slog.Int64("chatid", -1)
		}(),
	)
	templateData := h.defaultData(update)
	text := templates.Render("error", &templateData)

	if message != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    message.Chat.ID,
			Text:      text,
			ParseMode: models.ParseModeHTML,
		})
	}
}

func (h *Handler) defaultData(update *models.Update) map[string]any {
	data := map[string]any{
		"time": time.Now(),
	}
	message := getMessage(update)
	if message != nil {
		data["chatid"] = message.Chat.ID
		data["userid"] = message.Chat.ID
		data["first_name_tg"] = message.Chat.FirstName
		data["last_name_tg"] = message.Chat.LastName
	}
	return data
}
