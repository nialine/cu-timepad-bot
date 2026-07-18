package service

import (
	"context"
	"cu-timepad-bot/internal/config"
	"cu-timepad-bot/internal/domain"
	"cu-timepad-bot/pkg/timepad"
	"log/slog"
	"slices"
	"time"

	"github.com/patrickmn/go-cache"
	"golang.org/x/sync/errgroup"
)

func (h *Service) StartTimepadWorker(ctx context.Context, client *timepad.Client) {
	cfg := config.GetConfig(ctx)

	interval := time.Duration(cfg.TimepadFetchInterval) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			slog.LogAttrs(ctx,
				slog.LevelDebug,
				"Processing timepad events",
			)
			h.processTimepadEvents(ctx, client)
			slog.LogAttrs(ctx,
				slog.LevelDebug,
				"Stopped processing events",
			)
		}
	}
}

func (h *Service) processTimepadEvents(ctx context.Context, client *timepad.Client) {
	cfg := config.GetConfig(ctx)

	g, gctx := errgroup.WithContext(ctx)

	for _, ev := range cfg.Events {
		g.Go(func() error {
			err := h.processEvent(gctx, client, ev)
			if err != nil {
				slog.LogAttrs(ctx,
					slog.LevelDebug,
					"Event processed",
					slog.Int("eventid", int(ev.ID)),
				)
			}
			return err
		})
	}

	if err := g.Wait(); err != nil {
		slog.LogAttrs(ctx,
			slog.LevelError,
			"Error processing events",
			slog.Any("error", err),
		)
	}
}

func (h *Service) processEvent(ctx context.Context, client *timepad.Client, ev domain.Event) error {
	event, err := client.GetData(ctx, ev.URL)
	if err != nil {
		return err
	}
	recurring_events := deleteIrrelevantRecurringEvents(event.RecurringEvents)

	c, ok := h.cacheEvent.Get(event.Name)
	if ok {
		last_recurring_events := c.([]timepad.RecurringEvent)

		diff := make([]timepad.RecurringEvent, 0, 16)
		for _, v := range recurring_events {
			if !slices.ContainsFunc(last_recurring_events, func(ev timepad.RecurringEvent) bool {
				return ev.ID == v.ID && ev.Unavailable == v.Unavailable
			}) && !v.Unavailable {
				diff = append(diff, v)
			}
		}

		if len(diff) > 0 {
			h.cacheEvent.Set(event.Name, recurring_events, cache.DefaultExpiration)
			go h.NotifyPeople(ctx, ev, diff)
		}
	} else {
		h.cacheEvent.Set(event.Name, recurring_events, cache.DefaultExpiration)
	}
	return nil
}

func deleteIrrelevantRecurringEvents(events []timepad.RecurringEvent) []timepad.RecurringEvent {
	return slices.DeleteFunc(events, func(ev timepad.RecurringEvent) bool {
		return ev.TicketsLeft == nil
	})
}
