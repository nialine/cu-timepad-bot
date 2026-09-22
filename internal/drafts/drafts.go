package drafts

import (
	"context"
	"cu-timepad-bot/internal/domain"
	"cu-timepad-bot/internal/store"
	"encoding/json"
	"fmt"
	"time"
)

const defaultTTL = 15 * time.Minute

type State int

const registrationPrefixKey = "draft:register:"
const tempRegistrationPrefixKey = "draft:temp_register:"

const (
	StateEmail State = iota
	StateFirstName
	StateLastName
	StateDone
)

type RegistrationDraft struct {
	State State
	domain.BookingUserData
}

type TempRegistrationDraft struct {
	State State
	domain.BookingUserData
	domain.ChooseSlotData
}

func makeRegistrationKey(userid int64, temp bool) string {
	if temp {
		return fmt.Sprintf("%s%d", tempRegistrationPrefixKey, userid)
	}
	return fmt.Sprintf("%s%d", registrationPrefixKey, userid)
}

func saveDraft[T any](ctx context.Context, st store.DraftStore, key string, draft T, ttl time.Duration) error {
	b, err := json.Marshal(draft)
	if err != nil {
		return err
	}
	st.Save(ctx, key, b, ttl)
	return nil
}

func getDraft[T any](ctx context.Context, st store.DraftStore, key string) (*T, error) {
	b, err := st.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	var draft *T
	err = json.Unmarshal(b, &draft)
	if err != nil {
		return nil, err
	}
	return draft, nil
}

func LoadTempBooking(ctx context.Context, draftStore store.DraftStore, userid int64) (*TempRegistrationDraft, error) {
	return getDraft[TempRegistrationDraft](ctx, draftStore, makeRegistrationKey(userid, true))
}

func SaveTempBooking(ctx context.Context, draftStore store.DraftStore, userid int64, draft *TempRegistrationDraft) {
	saveDraft[TempRegistrationDraft](ctx, draftStore, makeRegistrationKey(userid, true), *draft, defaultTTL)
}

func LoadBooking(ctx context.Context, draftStore store.DraftStore, userid int64) (*RegistrationDraft, error) {
	return getDraft[RegistrationDraft](ctx, draftStore, makeRegistrationKey(userid, false))
}

func SaveBooking(ctx context.Context, draftStore store.DraftStore, userid int64, draft *RegistrationDraft) {
	saveDraft[RegistrationDraft](ctx, draftStore, makeRegistrationKey(userid, false), *draft, defaultTTL)
}

func HasBooking(ctx context.Context, draftStore store.DraftStore, userid int64) domain.Status {
	draft, _ := getDraft[RegistrationDraft](ctx, draftStore, makeRegistrationKey(userid, false))
	tempdraft, _ := getDraft[TempRegistrationDraft](ctx, draftStore, makeRegistrationKey(userid, true))
	switch {
	case draft != nil && draft.State != StateDone:
		return domain.StatusRegistration
	case tempdraft != nil && tempdraft.State != StateDone:
		return domain.StatusTempRegistration
	}
	return domain.StatusNone
}
