package service

import (
	"context"
	"cu-timepad-bot/internal/adapters/timepad"
	"cu-timepad-bot/internal/domain"
	"cu-timepad-bot/internal/drafts"
	"cu-timepad-bot/internal/store"
	"log/slog"
	"net/mail"
	"slices"
	"strconv"
)

func isEmailValid(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

type Service struct {
	st      store.UserStore
	draftst store.DraftStore

	cacheEvent    map[int64][]timepad.RecurringEvent
	timepadClient *timepad.Client
	NewSlots      chan *domain.NewSlots
}

func New(st store.UserStore, draftst store.DraftStore, timepad_client *timepad.Client) *Service {
	cache_event := make(map[int64][]timepad.RecurringEvent)

	return &Service{
		st,
		draftst,
		cache_event,
		timepad_client,
		make(chan *domain.NewSlots, 4),
	}
}

func (svc *Service) AddUser(ctx context.Context, userid int64) error {
	return svc.st.AddUser(ctx, userid)
}

func (svc *Service) GetUser(ctx context.Context, userid int64) (*domain.User, error) {
	return svc.st.GetUser(ctx, userid)
}

func (svc *Service) IsSubscribedUser(ctx context.Context, userid int64, eventid int64) (bool, error) {
	return svc.st.IsSubscribedUser(ctx, userid, eventid)
}

func (svc *Service) FindUsersWithEvent(ctx context.Context, eventid int64) []*domain.User {
	return svc.st.FindUsersWithEvent(ctx, eventid)
}

func (svc *Service) AddUserSubscribedEvent(ctx context.Context, userid int64, eventid int64) error {
	return svc.st.AddUserSubscribedEvent(ctx, userid, eventid)
}

func (svc *Service) RemoveUserSubscribedEvent(ctx context.Context, userid int64, eventid int64) error {
	return svc.st.RemoveUserSubscribedEvent(ctx, userid, eventid)
}

func (svc *Service) HasRegistrationData(ctx context.Context, userid int64) bool {
	user, _ := svc.GetUser(ctx, userid)
	return user.BookingUserData != nil
}

func (svc *Service) GetTempRegistration(ctx context.Context, userid int64) (*drafts.TempRegistrationDraft, error) {
	return drafts.LoadTempBooking(ctx, svc.draftst, userid)
}

func (svc *Service) InRegistrationProcess(ctx context.Context, userid int64) domain.Status {
	return drafts.HasBooking(ctx, svc.draftst, userid)
}

func (svc *Service) ProcessEventCallback(ctx context.Context, userid int64, callbackData []string) error {
	eventid, err := strconv.ParseInt(callbackData[0], 10, 64)
	if err != nil {
		return err
	}
	user, err := svc.GetUser(ctx, userid)
	if err != nil {
		return err
	}
	status := ""
	if slices.Contains(user.SubscribedEvents, eventid) {
		err = svc.RemoveUserSubscribedEvent(ctx,
			userid,
			eventid,
		)
		status = "unsubscribed"
	} else {
		err = svc.AddUserSubscribedEvent(ctx,
			userid,
			eventid,
		)
		status = "subscribed"
	}
	if err != nil {
		return err
	}
	slog.LogAttrs(ctx,
		slog.LevelInfo,
		"User changed subscription to event",
		slog.Int64("userid", userid),
		slog.Int64("eventid", eventid),
		slog.String("status", status),
	)
	return nil
}

func (svc *Service) GenSubscribeEventButton(ctx context.Context, userid int64, ev *domain.Event) (string, error) {
	is_subscribed, err := svc.IsSubscribedUser(ctx, userid, ev.ID)
	if err != nil {
		return "", err
	}
	text_prefix := ""
	if is_subscribed {
		text_prefix = "✓ "
	}
	return text_prefix + ev.Name, nil
}

func (svc *Service) GetAvaliableSlots(ctx context.Context, ev *domain.Event) ([]timepad.RecurringEvent, error) {
	timepad_slots, ok := svc.cacheEvent[ev.ID]
	if !ok {
		err := svc.processEvent(ctx, ev)
		if err != nil {
			slog.LogAttrs(ctx,
				slog.LevelError,
				"Error processing events",
				slog.Any("error", err),
				slog.Int("eventid", int(ev.ID)),
			)
			return nil, err
		} else {
			timepad_slots, _ = svc.cacheEvent[ev.ID]
		}
	}

	return timepad_slots, nil
}
