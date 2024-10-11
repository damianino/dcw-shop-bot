package telegram_bot_framework

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type DialogTree struct {
	root *Page
}

type Page struct {
	id       string
	text     string
	buttons  []button
	backPage *Page
}

type button struct {
	callbackData    string
	text            string
	page            *Page
	action          *Action
	defaultResponse []tgbotapi.MessageConfig
}

func NewDialogTree(root Page) DialogTree {
	return DialogTree{
		root: &root,
	}
}

func NewPage(id, text string, backPage *Page) *Page {
	return &Page{
		id:       id,
		text:     text,
		buttons:  nil,
		backPage: backPage,
	}
}

func (p *Page) AddButton(text, callbackData string, redirectTo *Page, action *Action, defaultResponse []tgbotapi.MessageConfig) {
	p.buttons = append(p.buttons, button{
		callbackData:    callbackData,
		text:            text,
		page:            redirectTo,
		action:          action,
		defaultResponse: defaultResponse,
	})
}

func (p *Page) RemoveAllButtons() {
	p.buttons = nil
}

func (p *Page) BuildMessage() tgbotapi.Chattable {
	msg := tgbotapi.MessageConfig{Text: p.text}
	replyMarkup := make([][]tgbotapi.InlineKeyboardButton, 0)

	for _, btn := range p.buttons {
		replyMarkup = append(
			replyMarkup,
			[]tgbotapi.InlineKeyboardButton{{
				Text:         btn.text,
				CallbackData: &btn.callbackData,
			}})
	}

	if p.backPage != nil {
		replyMarkup = append(
			replyMarkup,
			[]tgbotapi.InlineKeyboardButton{{
				Text:         "👈Back",
				CallbackData: &p.backPage.id,
			}})
	}

	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(replyMarkup...)

	return msg
}
