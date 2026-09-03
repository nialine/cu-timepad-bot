package handler

import (
	"context"
	"cu-timepad-bot/internal/config"
	"cu-timepad-bot/internal/templates"
	"fmt"
	"strconv"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type chooseSlotData struct {
	eventid int64
	page    int
}

func parseChooseSlotCallback(callbackData []string) *chooseSlotData {
	if len(callbackData) >= 1 {
		var data chooseSlotData
		data.eventid, _ = strconv.ParseInt(callbackData[0], 10, 64)
		if len(callbackData) >= 2 {
			data.page, _ = strconv.Atoi(callbackData[1])
		} else {
			data.page = 0
		}
		return &data
	}
	return nil
}

func (h *Handler) chooseSlotCallback(ctx context.Context, b *bot.Bot, update *models.Update, data *chooseSlotData) {
	cfg := config.GetConfig(ctx)
	templateData := h.defaultData(update)
	eventid := data.eventid
	page := data.page

	event, err := cfg.GetEvent(eventid)
	if err != nil {
		h.writeError(ctx, b, update, err)
	}

	slots, err := h.svc.GetAvaliableSlots(ctx, event)
	if err != nil {
		h.writeError(ctx, b, update, err)
	}
	max_page := len(slots) / slotsOnPage

	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{},
	}

	for i := page * slotsOnPage; i < min((page+1)*slotsOnPage, len(slots)); i++ {
		slot := slots[i]
		kb.InlineKeyboard = append(kb.InlineKeyboard, []models.InlineKeyboardButton{{
			Text:         slot.DateH,
			CallbackData: fmt.Sprintf("%v:%v:%v:%v", reserveSlotCallback, eventid, page, slot.ID),
		}})
	}
	if max_page > 0 {
		kb.InlineKeyboard = append(kb.InlineKeyboard, []models.InlineKeyboardButton{})
		last_elem := &kb.InlineKeyboard[len(kb.InlineKeyboard)-1]
		if page > 0 {
			*last_elem = append(*last_elem, models.InlineKeyboardButton{
				Text:         "<-",
				CallbackData: fmt.Sprintf("%v:%v:%v", chooseSlotCallback, eventid, page-1),
			})
		}
		if page < max_page {
			*last_elem = append(*last_elem, models.InlineKeyboardButton{
				Text:         "->",
				CallbackData: fmt.Sprintf("%v:%v:%v", chooseSlotCallback, eventid, page+1),
			})
		}
	}

	kb.InlineKeyboard = append(kb.InlineKeyboard, []models.InlineKeyboardButton{{
		Text:         templates.Render("back_button", &templateData),
		CallbackData: chooseEventForBookingCallback,
	}, {
		Text:         templates.Render("home_button", &templateData),
		CallbackData: startCallback,
	}})

	if len(slots) > 0 {
		b.EditMessageText(ctx, &bot.EditMessageTextParams{
			ChatID:      update.CallbackQuery.Message.Message.Chat.ID,
			MessageID:   update.CallbackQuery.Message.Message.ID,
			Text:        templates.Render("choose_slot", &templateData),
			ParseMode:   models.ParseModeHTML,
			ReplyMarkup: kb,
		})
	} else {
		b.EditMessageText(ctx, &bot.EditMessageTextParams{
			ChatID:      update.CallbackQuery.Message.Message.Chat.ID,
			MessageID:   update.CallbackQuery.Message.Message.ID,
			Text:        templates.Render("choose_slot_empty", &templateData),
			ParseMode:   models.ParseModeHTML,
			ReplyMarkup: kb,
		})
	}
}
