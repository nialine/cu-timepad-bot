package handler

import (
	"bytes"
	"cu-timepad-bot/internal/domain"
	"text/template"
)

var BotTemplates = map[string]string{
	"start": "Бот не работает, ожидайте починки...",
	"error": "<b>Error happened at {{.time}}</b>, your chat id: {{.chatid}}\nPlease contact @niazya",
}

func renderTemplate(templateName string, data any) (string, error) {
	tmplText, exists := BotTemplates[templateName]
	if !exists {
		return "", domain.ErrTemplateNotFound
	}

	tmpl, err := template.New(templateName).Parse(tmplText)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
