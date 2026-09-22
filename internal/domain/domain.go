package domain

import (
	"cu-timepad-bot/internal/adapters/timepad"
	"errors"
	"time"
)

var (
	ErrUserNotFound      = errors.New("User not found")
	ErrEventNotFound     = errors.New("Event not found")
	ErrSlotNotFound      = errors.New("Slot not found")
	ErrSlotIsUnavailable = errors.New("Slot is unavailable")
	ErrDataNotSufficient = errors.New("UserData not sufficient")
	ErrDataIsInvalid     = errors.New("Data is invalid")
)

type Status int

const (
	StatusNone Status = iota
	StatusRegistration
	StatusTempRegistration
)

type User struct {
	ID int64 `bson:"id,required"`

	*BookingUserData `bson:"bookinguserdata,omitempty"`

	SubscribedEvents []int64 `bson:"subscribed_events,omitempty"`
	Slots            []Slot  `bson:"slots,omitempty"`
}

type BookingUserData struct {
	FirstName string `bson:"first_name,omitempty"`
	LastName  string `bson:"last_name,omitempty"`
	Email     string `bson:"email,omitempty"`
}

type ChooseSlotData struct {
	EventID int64
	SlotID  int64
	Page    int
}

type Event struct {
	ID     int64  `yaml:"id"`
	Name   string `yaml:"name"`
	Name_r string `yaml:"name_r"`
	URL    string `yaml:"url"`
	// I DON'T KNOW HOW TIMEPAD API WORKS BUT THIS MAGIC NUMBER FIXES EVERYTHING
	MagicStr string `yaml:"magic_string"`
}

type NewSlots struct {
	Event    *Event
	NewSlots []timepad.RecurringEvent
}

type Slot struct {
	ID   int64
	Time time.Time
}
