package service

import (
	"context"
	"cu-timepad-bot/internal/domain"
	"cu-timepad-bot/internal/templates"
	"cu-timepad-bot/pkg/timepad"
	"log/slog"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"golang.org/x/sync/errgroup"
)

func (h *Service) NotifyPeople(ctx context.Context, ev domain.Event, new_slots []timepad.RecurringEvent) {
	slog.LogAttrs(ctx,
		slog.LevelDebug,
		"New slots",
		slog.Int64("eventid", int64(ev.ID)),
		slog.Any("new_slots", new_slots),
	)
	new_events := domain.NewSlots{Event: &ev, NewSlots: new_slots}
	h.NewSlots <- new_events
}

func (h *Service) StartNotifyingWorker(ctx context.Context, b *bot.Bot) {
	for {
		select {
		case <-ctx.Done():
			return
		case new_events := <-h.NewSlots:
			slog.LogAttrs(ctx,
				slog.LevelDebug,
				"Proccessing notifying people",
			)
			h.processNotifyPeople(ctx, b, new_events)
			slog.LogAttrs(ctx,
				slog.LevelDebug,
				"Stopped notifying people",
			)
		}
	}
}

func (h *Service) processNotifyPeople(ctx context.Context, b *bot.Bot, new_events domain.NewSlots) {
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(2)

	users := h.st.FindUsersWithEvent(ctx, new_events.Event.ID)

	for _, user := range users {
		g.Go(func() error {
			err := h.sendNotification(gctx, b, user.ID, new_events)
			return err
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

func (h *Service) sendNotification(ctx context.Context, b *bot.Bot, userid int64, new_events domain.NewSlots) error {
	templateData := map[string]any{
		"userid":    userid,
		"new_slots": new_events,
		"time":      time.Now(),
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    userid,
		Text:      templates.Render("new_slots_notification", templateData),
		ParseMode: models.ParseModeHTML,
	})
	return nil
}
