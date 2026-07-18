package templates

import (
	"cu-timepad-bot/internal/domain"
	"cu-timepad-bot/pkg/timepad"
	"strings"
	"testing"
	"time"
)

func TestNewSlotsNotification(t *testing.T) {
	time0 := time.Now()
	time1 := time.Now().Add(time.Hour)
	tickets_left := 1
	new_slots := domain.NewSlots{
		Event: &domain.Event{
			ID:   0,
			Name: "Тестовый ивент",
			URL:  "http://example.com/",
		},
		NewSlots: []timepad.RecurringEvent{
			{
				ID: 1,

				Time:  time0,
				Date:  time0.Format(time.DateOnly + " " + time.TimeOnly),
				DateH: time0.Format("Monday, 1 January at 15:03"),

				TicketsLeft: &tickets_left,
				Unavailable: false,
				IsCurrent:   false,
			},
			{
				ID: 2,

				Time:  time1,
				Date:  time1.Format(time.DateOnly + " " + time.TimeOnly),
				DateH: time1.Format("Monday, 1 January at 15:03"),

				TicketsLeft: &tickets_left,
				Unavailable: false,
				IsCurrent:   false,
			},
		},
	}

	data := map[string]any{
		"new_slots": new_slots,
		"time":      time.Now(),
	}

	text := Render("new_slots_notification", data)
	t.Logf("Resulted test: %v", text)
	if strings.HasPrefix(text, "Err") {
		t.Error("new_slots_notification template returned error")
	}
}
