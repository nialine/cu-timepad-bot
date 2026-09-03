package service

import (
	"context"
	"cu-timepad-bot/internal/adapters/timepad"
	"cu-timepad-bot/internal/config"
	"cu-timepad-bot/internal/domain"
	"log/slog"
	"slices"
	"time"

	"golang.org/x/sync/errgroup"
)

func (svc *Service) getSlot(eventid, slotid int64) *timepad.RecurringEvent {
	timepad_slots, _ := svc.cacheEvent[eventid]

	for _, slot := range slices.Backward(timepad_slots) {
		if slot.ID == slotid {
			return &slot
		}
	}
	return nil
}

func (svc *Service) StartTimepadWorker(ctx context.Context) {
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
			svc.processTimepadEvents(ctx)
			slog.LogAttrs(ctx,
				slog.LevelDebug,
				"Stopped processing events",
			)
		}
	}
}

func (svc *Service) processTimepadEvents(ctx context.Context) {
	cfg := config.GetConfig(ctx)

	g, gctx := errgroup.WithContext(ctx)

	for _, ev := range cfg.Events {
		g.Go(func() error {
			err := svc.processEvent(gctx, ev)
			if err == nil {
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

func (svc *Service) processEvent(ctx context.Context, ev *domain.Event) error {
	var event *timepad.Event
	var err error
	if ev.URL == "" {
		event, err = svc.timepadClient.GetEventDataID(ctx, ev.ID)
	} else {
		event, err = svc.timepadClient.GetEventDataURL(ctx, ev.URL)
	}
	if err != nil {
		return err
	}

	recurring_events := deleteUnavailableRecurringEvents(event.RecurringEvents)

	last_recurring_events, ok := svc.cacheEvent[ev.ID]
	if ok {
		diff := make([]timepad.RecurringEvent, 0, 16)
		for _, v := range recurring_events {
			if !slices.ContainsFunc(last_recurring_events, func(ev timepad.RecurringEvent) bool {
				return ev.ID == v.ID && ev.Unavailable == v.Unavailable
			}) {
				diff = append(diff, v)
			}
		}

		if len(diff) > 0 {
			svc.cacheEvent[ev.ID] = recurring_events
			go svc.NotifyPeople(ctx, ev, diff)
		}
	} else {
		svc.cacheEvent[ev.ID] = recurring_events
	}
	return nil
}

func deleteIrrelevantRecurringEvents(events []timepad.RecurringEvent) []timepad.RecurringEvent {
	return slices.DeleteFunc(events, func(ev timepad.RecurringEvent) bool {
		return ev.TicketsLeft == nil
	})
}

func deleteUnavailableRecurringEvents(events []timepad.RecurringEvent) []timepad.RecurringEvent {
	return slices.DeleteFunc(events, func(ev timepad.RecurringEvent) bool {
		return ev.TicketsLeft == nil || *ev.TicketsLeft == 0
	})
}
