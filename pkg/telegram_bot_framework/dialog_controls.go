package telegram_bot_framework

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

type DialogControls struct {
	UpdatesIn chan tgbotapi.Update
	MsgOut    chan tgbotapi.Chattable
}

func (dc *DialogControls) Prompt(prompt Prompt) string {
	msg := tgbotapi.NewMessage(0, prompt.params.Text)
	switch prompt.params.AnsType {
	case AnsTypeText:
		break
	case AnsTypeVariant:
		keyboardMarkup := tgbotapi.NewInlineKeyboardMarkup()
		for i, v := range prompt.params.Variants {
			keyboardMarkup.InlineKeyboard = append(
				keyboardMarkup.InlineKeyboard,
				tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(v, fmt.Sprintf("%s:%d", prompt.id, i))),
			)
		}
		msg.ReplyMarkup = keyboardMarkup
	default:
		log.Printf("unknown AnsType in pormpt params: %d", prompt.params.AnsType)
	}

	var errReply string
	var success bool
	for {
		if success {
			break
		}

		dc.MsgOut <- msg

		select {
		case update := <-dc.UpdatesIn:
			switch prompt.params.AnsType {
			case AnsTypeText:
				errReply, success = prompt.Validate(update.Message.Text)
				if success {
					return update.Message.Text
				}
				break
			case AnsTypeVariant:
				log.Printf("impliment me")
			}
		}
		dc.MsgOut <- tgbotapi.NewMessage(0, errReply)
	}

	return ""
}
