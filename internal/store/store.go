package store

import (
	"context"
	"cu-timepad-bot/internal/domain"
)

type Store interface {
	AddUser(ctx context.Context, userid int64) error
	GetUser(ctx context.Context, userid int64) (*domain.User, error)
	IsSubscribedUser(ctx context.Context, userid int64, eventid domain.EventID) (bool, error)
	FindUsersWithEvent(ctx context.Context, eventid domain.EventID) []domain.User
	AddUserSubscribedEvent(ctx context.Context, userid int64, eventid domain.EventID) error
	RemoveUserSubscribedEvent(ctx context.Context, userid int64, eventid domain.EventID) error
}
