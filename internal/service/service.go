package service

import (
	"context"
	"cu-timepad-bot/internal/domain"
	"cu-timepad-bot/internal/store"
	"log/slog"
	"slices"
	"strconv"
	"time"

	"github.com/patrickmn/go-cache"
)

type Service struct {
	st         store.Store
	cacheEvent *cache.Cache

	NewSlots chan domain.NewSlots
}

func New(st store.Store) *Service {
	cache_event := cache.New(cache.NoExpiration, time.Minute)

	return &Service{
		st,
		cache_event,
		make(chan domain.NewSlots, 4),
	}
}

func (s *Service) AddUser(ctx context.Context, userid int64) error {
	return s.st.AddUser(ctx, userid)
}

func (s *Service) GetUser(ctx context.Context, userid int64) (*domain.User, error) {
	return s.st.GetUser(ctx, userid)
}

func (s *Service) IsSubscribedUser(ctx context.Context, userid int64, eventid domain.EventID) (bool, error) {
	return s.st.IsSubscribedUser(ctx, userid, eventid)
}

func (s *Service) FindUsersWithEvent(ctx context.Context, eventid domain.EventID) []domain.User {
	return s.st.FindUsersWithEvent(ctx, eventid)
}

func (s *Service) AddUserSubscribedEvent(ctx context.Context, userid int64, eventid domain.EventID) error {
	return s.st.AddUserSubscribedEvent(ctx, userid, eventid)
}

func (s *Service) RemoveUserSubscribedEvent(ctx context.Context, userid int64, eventid domain.EventID) error {
	return s.st.RemoveUserSubscribedEvent(ctx, userid, eventid)
}

func (s *Service) ProcessEventCallback(ctx context.Context, userid int64, callbackData []string) error {
	eventid, err := strconv.ParseInt(callbackData[1], 10, 64)
	if err != nil {
		return err
	}
	user, err := s.GetUser(ctx, userid)
	if err != nil {
		return err
	}
	status := ""
	if slices.Contains(user.SubscribedEvents, domain.EventID(eventid)) {
		err = s.RemoveUserSubscribedEvent(ctx,
			userid,
			domain.EventID(eventid),
		)
		status = "unsubscribed"
	} else {
		err = s.AddUserSubscribedEvent(ctx,
			userid,
			domain.EventID(eventid),
		)
		status = "subscribed"
	}
	if err != nil {
		return err
	}
	slog.LogAttrs(ctx,
		slog.LevelDebug,
		"User changed subscription to event",
		slog.Int64("userid", userid),
		slog.Int64("eventid", eventid),
		slog.String("status", status),
	)
	return nil
}

func (s *Service) GenEventButton(ctx context.Context, userid int64, ev domain.Event) (string, error) {
	is_subscribed, err := s.IsSubscribedUser(ctx, userid, ev.ID)
	if err != nil {
		return "", err
	}
	text_prefix := ""
	if is_subscribed {
		text_prefix = "✓ "
	}
	return text_prefix + ev.Name, nil
}
