package timepad

import (
	"errors"
	"time"
)

var (
	ErrParse = errors.New("Can't parse event")
)

type Event struct {
	ID              int              `json:"id"`
	Name            string           `json:"name"`
	IsRecurring     bool             `json:"is_recurring"`
	RecurringEvents []RecurringEvent `json:"recurring_events"`
}

type RecurringEvent struct {
	ID   int       `json:"id"`
	Time time.Time `json:"date_iso8601"`

	Date  string `json:"date"`
	DateH string `json:"date_h"`

	TicketsLeft *int `json:"tickets_left"`
	Unavailable bool `json:"unavaliable"`
	IsCurrent   bool `json:"isCurrent"`
}
