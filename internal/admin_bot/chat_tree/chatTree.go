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
	startPageID    = "startPage"
	selectDialogCB = "selectDialog"

	DialogTypeGeneral = "general"
)

type Deps struct {
	Repo *repository.Repository
}

func InitChatTree(deps *Deps) tgbf.DialogTree {
	startPage := tgbf.NewPage(startPageID, "Start Page", nil)

	dialogsPage := tgbf.NewPage(startPageID, "Dialogs", startPage)

	startPage.AddButton(
		"Dialogs",
		"dialogs",
		dialogsPage,
		tgbf.NewAction(
			"updateDialogsList",
			nil,
			func(ctx context.Context, controls tgbf.DialogControls) error {
				return UpdateChatTreeDialogsList(ctx, deps, dialogsPage)
			},
			startPage,
			startPage,
		),
		nil,
	)

	return tgbf.NewDialogTree(*startPage)
}

func UpdateChatTreeDialogsList(ctx context.Context, deps *Deps, dialogsPage *tgbf.Page) error {
	dialogsPage.RemoveAllButtons()

	dialogs, err := deps.Repo.Dialogs.GetAllUncompleted()
	if err != nil {
		return err
	}

	for _, d := range dialogs {
		dialogsPage.AddButton(
			d.TGClientUsername,
			fmt.Sprintf("%s:%d", selectDialogCB, d.ID),
			nil,
			tgbf.NewAction(
				"enterLiveChat",
				nil,
				func(ctx context.Context, controls tgbf.DialogControls) error {
					return EnterLiveChat(ctx, d, controls, deps)
				},
				dialogsPage,
				dialogsPage,
			),
			nil,
		)
	}

	return nil
}

func EnterLiveChat(ctx context.Context, dialog models.Dialog, controls tgbf.DialogControls, deps *Deps) error {
	messages, err := deps.Repo.Messages.GetAllByDialogID(dialog.ID)
	if err != nil {
		log.Println("failed to get dialog messages")
	}

	refreshIncomingTicker := time.NewTicker(time.Second * 5)
	deliveredMessageIDs := make(map[int64]struct{})

	for _, m := range messages {
		var text string

		if m.Text != nil {
			text = fmt.Sprintf("\n\n%s", *m.Text)
		}

		if m.SentByCustomer {
			text = fmt.Sprintf("*%s:*%s", dialog.TGClientUsername, tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, text))
		} else {
			text = fmt.Sprintf("*You:*%s", tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, text))
		}

		deliveredMessageIDs[m.ID] = struct{}{}

		msg := tgbotapi.NewMessage(0, text)
		msg.ParseMode = tgbotapi.ModeMarkdown
		controls.MsgOut <- msg
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("stopping EnterLiveChat, ctx done")
			return nil

		case update := <-controls.UpdatesIn:
			switch {
			// command
			case update.Message != nil && update.Message.From != nil && update.Message.IsCommand():
				switch update.Message.Command() {
				case "back":
					return nil
				case "close":
					if err := deps.Repo.Dialogs.MarkCompleted(&dialog); err != nil {
						controls.MsgOut <- tgbotapi.NewMessage(0, "Error while marking dialog as completed")
						return err
					}

					controls.MsgOut <- tgbotapi.NewMessage(
						0,
						fmt.Sprintf("Dialog with %s marked as completed", dialog.TGClientUsername),
					)
					return err
				}
			// message
			case update.Message != nil && update.Message.From != nil:
				msg := &models.Message{
					DialogID:       dialog.ID,
					SentByCustomer: false,
					Text:           &(update.Message.Text),
				}
				if err := deps.Repo.Messages.Create(msg); err != nil {
					log.Println(err)
				}
				deliveredMessageIDs[msg.ID] = struct{}{}
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
