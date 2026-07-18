package handler

import (
	"context"
	"cu-timepad-bot/internal/config"
	"cu-timepad-bot/internal/domain"
	"cu-timepad-bot/internal/templates"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/patrickmn/go-cache"
)

type service interface {
	AddUser(ctx context.Context, userid int64) error
	GetUser(ctx context.Context, userid int64) (*domain.User, error)
	ProcessEventCallback(ctx context.Context, userid int64, callbackData []string) error
	GenEventButton(ctx context.Context, userid int64, ev domain.Event) (string, error)
}

type Handler struct {
	svc        service
	cacheEvent *cache.Cache

	NewSlots chan domain.NewSlots
}

func New(st service) Handler {
	cache_event := cache.New(cache.NoExpiration, time.Minute)

	return Handler{
		st,
		cache_event,
		make(chan domain.NewSlots, 4),
	}
}

type callbacks string

const (
	eventsCallback = "events"
	startCallback  = "start"
)

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
	text := templates.Render("error", templateData)

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

func (h *Handler) CallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       false,
	})

	data := strings.Split(update.CallbackQuery.Data, ":")
	switch data[0] {
	case eventsCallback:
		h.ShowEventsCallback(ctx, b, update, data)
	case startCallback:
		h.EditStart(ctx, b, update)
	default:
		slog.LogAttrs(ctx,
			slog.LevelWarn,
			"Unknown callback",
			slog.String("callback", update.CallbackQuery.Data),
		)
	}
}

func (h *Handler) ShowEventsCallback(ctx context.Context, b *bot.Bot, update *models.Update, callbackData []string) {
	templateData := h.defaultData(update)
	cfg := config.GetConfig(ctx)
	userid := update.CallbackQuery.Message.Message.Chat.ID

	if len(callbackData) > 1 {
		if err := h.svc.ProcessEventCallback(ctx, userid, callbackData); err != nil {
			h.writeError(ctx, b, update, err)
			return
		}
	}

	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{},
	}

	for _, ev := range cfg.Events {
		text, err := h.svc.GenEventButton(ctx, userid, ev)
		if err != nil {
			h.writeError(ctx, b, update, err)
			return
		}
		kb.InlineKeyboard = append(kb.InlineKeyboard, []models.InlineKeyboardButton{{
			Text:         text,
			CallbackData: fmt.Sprintf("%v%v%v", eventsCallback, ":", ev.ID),
		}})
	}

	kb.InlineKeyboard = append(kb.InlineKeyboard, []models.InlineKeyboardButton{{
		Text:         templates.Render("back_button", templateData),
		CallbackData: startCallback,
	}})

	_, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      userid,
		MessageID:   update.CallbackQuery.Message.Message.ID,
		Text:        templates.Render("show_events", templateData),
		ReplyMarkup: kb,
	})
	if err != nil {
		h.writeError(ctx, b, update, err)
	}
}

func (h *Handler) EditStart(ctx context.Context, b *bot.Bot, update *models.Update) {
	message := getMessage(update)

	templateData := h.defaultData(update)

	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{
					Text:         templates.Render("events_button", templateData),
					CallbackData: eventsCallback,
				},
			},
		},
	}

	b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      message.Chat.ID,
		MessageID:   message.ID,
		Text:        templates.Render("start", templateData),
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

	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{
					Text:         templates.Render("events_button", templateData),
					CallbackData: eventsCallback,
				},
			},
		},
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      userid,
		Text:        templates.Render("start", templateData),
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: kb,
	})
}
