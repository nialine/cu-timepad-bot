package service

import (
	"context"
	"cu-timepad-bot/internal/adapters/timepad"
	"cu-timepad-bot/internal/config"
	"cu-timepad-bot/internal/domain"
	"cu-timepad-bot/internal/drafts"
	"log/slog"
	"slices"
)

func (svc *Service) ReserveSlot(ctx context.Context, userid int64, eventid int64, slotid int64, userdata *domain.BookingUserData) (*timepad.RecurringEvent, error) {
	cfg := config.GetConfig(ctx)
	ev, _ := cfg.GetEvent(eventid)
	if userdata == nil {
		user, err := svc.GetUser(ctx, userid)
		if err != nil {
			return nil, err
		}
		if user.BookingUserData == nil {
			return nil, domain.ErrDataNotSufficient
		}
		userdata = user.BookingUserData
	}
	timepad_slots, _ := svc.cacheEvent[ev.ID]

	for _, slot := range slices.Backward(timepad_slots) {
		if slot.ID == slotid && slot.Unavailable {
			return nil, domain.ErrSlotIsUnavailable
		} else if slot.ID == slotid {
			_, err := svc.timepadClient.ReserveSlot(ctx, ev.MagicStr, ev.ID, slotid, userdata.Email, userdata.LastName, userdata.FirstName)
			if err != nil {
				return nil, err
			}
			return &slot, nil
		}
	}
	return nil, domain.ErrSlotNotFound
}

func (svc *Service) HandleRegistrationInput(ctx context.Context, userid int64, text string) (drafts.State, error) {
	draft, err := drafts.LoadBooking(ctx, svc.draftst, userid)
	if err != nil {
		draft = &drafts.RegistrationDraft{}
		drafts.SaveBooking(ctx, svc.draftst, userid, draft)
	}

	fall := false
	switch draft.State {
	case drafts.StateEmail:
		if text != "" {
			if !isEmailValid(text) {
				return draft.State, domain.ErrDataIsInvalid
			}
			draft.Email = text
			draft.State = drafts.StateFirstName
			drafts.SaveBooking(ctx, svc.draftst, userid, draft)
			fall = true
		} else {
			return draft.State, nil
		}
		fallthrough
	case drafts.StateFirstName:
		if !fall && text != "" {
			draft.FirstName = text
			draft.State = drafts.StateLastName
			drafts.SaveBooking(ctx, svc.draftst, userid, draft)
			fall = true
		} else {
			return draft.State, nil
		}
		fallthrough
	case drafts.StateLastName:
		if !fall && text != "" {
			draft.LastName = text
			draft.State = drafts.StateDone
			drafts.SaveBooking(ctx, svc.draftst, userid, draft)
		} else {
			return draft.State, nil
		}
	}

	slog.LogAttrs(ctx,
		slog.LevelDebug,
		"AddingUserData",
		slog.Any("userdata", draft.BookingUserData),
	)
	err = svc.st.AddUserBookingData(ctx, userid, &draft.BookingUserData)
	return draft.State, err
}

func (svc *Service) HandleTempBookingInput(ctx context.Context, userid int64, text string, data *domain.ChooseSlotData) (drafts.State, error) {
	draft, err := drafts.LoadTempBooking(ctx, svc.draftst, userid)
	if err != nil || draft.State == drafts.StateDone {
		draft = &drafts.TempRegistrationDraft{}
		drafts.SaveTempBooking(ctx, svc.draftst, userid, draft)
	}

	if data != nil {
		draft.ChooseSlotData = *data
		drafts.SaveTempBooking(ctx, svc.draftst, userid, draft)
	}

	fall := false
	switch draft.State {
	case drafts.StateEmail:
		if text != "" {
			if !isEmailValid(text) {
				return draft.State, domain.ErrDataIsInvalid
			}
			draft.Email = text
			draft.State = drafts.StateFirstName
			drafts.SaveTempBooking(ctx, svc.draftst, userid, draft)
			fall = true
		} else {
			return draft.State, nil
		}
		fallthrough
	case drafts.StateFirstName:
		if !fall && text != "" {
			draft.FirstName = text
			draft.State = drafts.StateLastName
			drafts.SaveTempBooking(ctx, svc.draftst, userid, draft)
			fall = true
		} else {
			return draft.State, nil
		}
		fallthrough
	case drafts.StateLastName:
		if !fall && text != "" {
			draft.LastName = text
			draft.State = drafts.StateDone
		} else {
			return draft.State, nil
		}
	}
	drafts.SaveTempBooking(ctx, svc.draftst, userid, draft)
	return draft.State, nil
}
