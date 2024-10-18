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

	msg := tgbotapi.NewMessage(0, "started live chat")
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Case closed")))
	controls.MsgOut <- msg

	refreshIncomingTicker := time.NewTicker(time.Second * 5)
	deliveredMessageIDs := make(map[int64]struct{})
	var messageWithMediaGroup struct {
		messageID    int64
		mediaGroupID string
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("ctx done, stopping StartLiveChat")
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

				if update.Message.MediaGroupID != "" {
					if messageWithMediaGroup.mediaGroupID != update.Message.MediaGroupID {
						messageWithMediaGroup.mediaGroupID = update.Message.MediaGroupID
						messageWithMediaGroup.messageID = msg.ID
					}

					var fileID string
					var mediaType string
					switch {
					case len(update.Message.Photo) != 0:
						fileID = update.Message.Photo[len(update.Message.Photo)-1].FileID
						mediaType = models.MediaTypePhoto
					case update.Message.Video != nil:
						fileID = update.Message.Video.FileID
						mediaType = models.MediaTypeVideo
					case update.Message.Document != nil:
						fileID = update.Message.Document.FileID
						mediaType = models.MediaTypeDocument
					default:

					}

					if err := deps.Repo.Media.Create(&models.Media{
						MessageID: messageWithMediaGroup.messageID,
						TGFileID:  fileID,
						Type:      mediaType,
					}); err != nil {
						log.Println(err)
					}
				}

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
					if m.Text != nil {
						controls.MsgOut <- tgbotapi.NewMessage(0, *m.Text)
					}
					media, err := deps.Repo.Media.GetAllByMessageID(m.ID)
					if err != nil {
						println(err)
					}
					if mediaOut := getMediaOut(media); mediaOut != nil {
						controls.MsgOut <- mediaOut
					}

					deliveredMessageIDs[m.ID] = struct{}{}
				}
			}
		}
	}
}

func getMediaOut(media []models.Media) interface{} {
	switch {
	case len(media) == 1:
		switch media[0].Type {
		case models.MediaTypePhoto:
			return tgbotapi.NewPhoto(0, tgbotapi.FileID(media[0].TGFileID))
		case models.MediaTypeVideo:
			return tgbotapi.NewVideo(0, tgbotapi.FileID(media[0].TGFileID))
		case models.MediaTypeDocument:
			return tgbotapi.NewDocument(0, tgbotapi.FileID(media[0].TGFileID))
		}

	case len(media) > 1:
		var mediaGroup []interface{}
		for _, m := range media {
			mediaGroup = append(mediaGroup, tgbotapi.FileID(m.TGFileID))
		}
		return tgbotapi.NewMediaGroup(0, mediaGroup)
	}

	return nil
}
