package domain

import (
	"cu-timepad-bot/pkg/timepad"
	"errors"
	"time"
)

var (
	ErrUserNotFound = errors.New("User not found")
)

type EventID int

type User struct {
	ID int64 `bson:"id,required"`

	FirstName string `bson:"first_name,omitempty"`
	LastName  string `bson:"last_name,omitempty"`
	email     string `bson:"email,omitempty"`

	SubscribedEvents []EventID `bson:"subscribed_events,omitempty"`
	Slots            []Slot    `bson:"slots,omitempty"`
}

type Event struct {
	ID   EventID `yaml:"id"`
	Name string  `yaml:"name"`
	URL  string  `yaml:"url"`
}

type NewSlots struct {
	Event    *Event
	NewSlots []timepad.RecurringEvent
}

type Slot struct {
	ID   int64
	Time time.Time
}
