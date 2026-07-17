package domain

import (
	"errors"
	"time"
)

var (
	ErrUserNotFound     = errors.New("User not found")
	ErrTemplateNotFound = errors.New("No template")
)

type EventID int

type User struct {
	ID int64

	FirstName string `bson:"first_name,omitempty"`
	LastName  string `bson:"last_name,omitempty"`
	email     string `bson:"email,omitempty"`

	SubscribedEvents []EventID `bson:"subscribed_events,omitempty"`
	Slots            []Slot    `bson:"slots,omitempty"`
}

type Slot struct {
	ID   int64
	Time time.Time
}
