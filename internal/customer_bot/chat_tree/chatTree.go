package chat_tree

import (
	"context"
	"fmt"
	"github.com/damianino/dcw-shop-bot/internal/repository"
	"github.com/damianino/dcw-shop-bot/internal/repository/models"
	tgbf "github.com/damianino/dcw-shop-bot/pkg/telegram_bot_framework"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"time"
)

const (
	startPageID   = "startPage"
	supportPageID = "supportPage"

	DialogTypeGeneral = "general"
)

type Deps struct {
	Repo *repository.Repository
}

func InitChatTree(deps *Deps) tgbf.DialogTree {
	startPage := tgbf.NewPage(startPageID, "start page!", nil)
	supportPage := tgbf.NewPage(supportPageID, "support!", startPage)

	startPage.AddButton(
		"support",
		"support",
		supportPage,
		nil,
		nil,
	)

	supportPage.AddButton(
		"How to return my order?",
		"return_order",
		nil,
		nil,
		[]tgbotapi.MessageConfig{
			{Text: "to return your order bla bla"},
		},
	)
	supportPage.AddButton(
		"Order question",
		"order_question",
		nil,
		tgbf.NewAction(
			"startLiveChat",
			nil,
			func(ctx context.Context, controls tgbf.DialogControls) error {
				return StartLiveChat(ctx, controls, deps)
			},
			startPage,
			startPage,
		),
		nil,
	)

	return tgbf.NewDialogTree(*startPage)
}

func StartLiveChat(ctx context.Context, controls tgbf.DialogControls, deps *Deps) error {
	dd := tgbf.GetContextDialogDetails(ctx)
	dialog := &models.Dialog{
		TGChatID:         dd.ChatID,
		TGUserID:         dd.UserID,
		TGClientUsername: dd.Username,
		DialogType:       DialogTypeGeneral,
		CompletedAt:      nil,
	}
	if err := deps.Repo.Dialogs.Create(dialog); err != nil {
		log.Println(err)
		return fmt.Errorf("server error")
	}

	msg := tgbotapi.NewMessage(0, "ok")
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Case closed")))
	controls.MsgOut <- msg

	refreshIncomingTicker := time.NewTicker(time.Second * 5)
	deliveredMessageIDs := make(map[int64]struct{})

	for {
		select {
		case <-ctx.Done():
			log.Println("stopping StartLiveChat, ctx done")
			return nil

		case update := <-controls.UpdatesIn:
			switch {
			// command
			case update.Message != nil && update.Message.From != nil && update.Message.IsCommand():

			// message
			case update.Message != nil && update.Message.From != nil:
				msg := &models.Message{
					DialogID:       dialog.ID,
					SentByCustomer: true,
					Text:           &(update.Message.Text),
				}
				if err := deps.Repo.Messages.Create(msg); err != nil {
					log.Println(err)
				}
				deliveredMessageIDs[msg.ID] = struct{}{}

				if update.Message.Text == "Case closed" {
					msg := tgbotapi.NewMessage(0, "Thank you!")
					msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)
					controls.MsgOut <- msg

					return deps.Repo.Dialogs.MarkCompleted(dialog)
				}
			// callback
			case update.CallbackQuery != nil:
			}

		case <-refreshIncomingTicker.C:
			messages, err := deps.Repo.Messages.GetAllByDialogID(dialog.ID)
			if err != nil {
				log.Println(err)
				continue
			}

			for _, m := range messages {
				if _, ok := deliveredMessageIDs[m.ID]; !ok {
					deliveredMessageIDs[m.ID] = struct{}{}
					controls.MsgOut <- tgbotapi.NewMessage(0, *m.Text)
				}
			}
		}
	}
}
