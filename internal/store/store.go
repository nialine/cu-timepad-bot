package store

import (
	"context"
	"cu-timepad-bot/internal/domain"
	"time"
)

type UserStore interface {
	AddUser(ctx context.Context, userid int64) error
	GetUser(ctx context.Context, userid int64) (*domain.User, error)
	IsSubscribedUser(ctx context.Context, userid int64, eventid int64) (bool, error)
	FindUsersWithEvent(ctx context.Context, eventid int64) []*domain.User
	AddUserSubscribedEvent(ctx context.Context, userid int64, eventid int64) error
	RemoveUserSubscribedEvent(ctx context.Context, userid int64, eventid int64) error
	AddUserBookingData(ctx context.Context, userid int64, data *domain.BookingUserData) error
}

type DraftStore interface {
	Save(ctx context.Context, key string, val []byte, ttl time.Duration) error
	Get(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) bool
}
