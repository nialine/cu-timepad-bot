package service

import (
	"context"
	"cu-timepad-bot/internal/adapters/timepad"
	"cu-timepad-bot/internal/domain"
	"cu-timepad-bot/internal/handler"
	"cu-timepad-bot/internal/templates"
	"log/slog"
	"strconv"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"golang.org/x/sync/errgroup"
)

func (svc *Service) NotifyPeople(ctx context.Context, ev *domain.Event, new_slots []timepad.RecurringEvent) {
	slog.LogAttrs(ctx,
		slog.LevelDebug,
		"New slots",
		slog.Int64("eventid", int64(ev.ID)),
		slog.Any("new_slots", new_slots),
	)
	new_events := domain.NewSlots{Event: ev, NewSlots: new_slots}
	svc.NewSlots <- &new_events
}

func (svc *Service) StartNotifyingWorker(ctx context.Context, b *bot.Bot) {
	for {
		select {
		case <-ctx.Done():
			return
		case new_events := <-svc.NewSlots:
			slog.LogAttrs(ctx,
				slog.LevelDebug,
				"Proccessing notifying people",
			)
			svc.processNotifyPeople(ctx, b, new_events)
			slog.LogAttrs(ctx,
				slog.LevelDebug,
				"Stopped notifying people",
			)
		}
	}
}

func (svc *Service) processNotifyPeople(ctx context.Context, b *bot.Bot, new_events *domain.NewSlots) {
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(2)

	users := svc.st.FindUsersWithEvent(ctx, new_events.Event.ID)

	for _, user := range users {
		user := user
		g.Go(func() error {
			slog.LogAttrs(ctx,
				slog.LevelDebug,
				"Notifying",
				slog.Int64("userid", user.ID),
			)
			return svc.sendNotification(gctx, b, user.ID, new_events)
		})
	}

	if err := g.Wait(); err != nil {
		slog.LogAttrs(ctx,
			slog.LevelError,
			"Error sending notification",
			slog.Any("error", err),
		)
	}
}

func (svc *Service) sendNotification(ctx context.Context, b *bot.Bot, userid int64, new_events *domain.NewSlots) error {
	templateData := map[string]any{
		"userid":    userid,
		"new_slots": new_events,
		"time":      time.Now(),
	}

	kb := models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{{
				Text: templates.Render("open_event_button", &templateData),
				URL:  new_events.Event.URL,
			}},
			{{
				Text:         templates.Render("reserve_slot_button", &templateData),
				CallbackData: handler.ChooseSlotCallback + ":" + strconv.FormatInt(new_events.Event.ID, 10),
			}}},
	}

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      userid,
		Text:        templates.Render("new_slots_notification", &templateData),
		ReplyMarkup: kb,
		ParseMode:   models.ParseModeHTML,
	})
	return err
}
