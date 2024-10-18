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

		media, err := deps.Repo.Media.GetAllByMessageID(m.ID)
		if err != nil {
			println(err)
		}
		if mediaOut := getMediaOut(dialog.TGChatID, media); mediaOut != nil {
			controls.MsgOut <- mediaOut
		}
	}

	refreshIncomingTicker := time.NewTicker(time.Second * 5)
	var messageWithMediaGroup struct {
		messageID    int64
		mediaGroupID string
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
				if update.Message.MediaGroupID != "" {
					if messageWithMediaGroup.mediaGroupID != update.Message.MediaGroupID {
						msg := &models.Message{
							DialogID:       dialog.ID,
							SentByCustomer: false,
							Text:           &(update.Message.Text),
						}
						if err := deps.Repo.Messages.Create(msg); err != nil {
							log.Println(err)
						}

						deliveredMessageIDs[msg.ID] = struct{}{}
						messageWithMediaGroup.mediaGroupID = update.Message.MediaGroupID
						messageWithMediaGroup.messageID = msg.ID
					}

					var fileID string
					var mediaType string
					switch {
					case len(update.Message.Photo) == 0:
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
					break
				}

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
					if m.Text != nil {
						controls.MsgOut <- tgbotapi.NewMessage(0, *m.Text)
					}
					media, err := deps.Repo.Media.GetAllByMessageID(m.ID)
					if err != nil {
						println(err)
					}
					if mediaOut := getMediaOut(dialog.TGChatID, media); mediaOut != nil {
						controls.MsgOut <- mediaOut
					}

					deliveredMessageIDs[m.ID] = struct{}{}
				}
			}
		}
	}
}

func getMediaOut(chatID int64, media []models.Media) interface{} {
	switch {
	case len(media) == 1:
		switch {
		case media[0].IsPhoto():
			return tgbotapi.NewPhoto(chatID, tgbotapi.FileID(media[0].TGFileID))
		case media[0].IsVideo():
			return tgbotapi.NewVideo(chatID, tgbotapi.FileID(media[0].TGFileID))
		case media[0].IsDocument():
			return tgbotapi.NewDocument(chatID, tgbotapi.FileID(media[0].TGFileID))
		}

	case len(media) > 1:
		var mediaGroup []interface{}
		for _, m := range media {
			switch {
			case m.IsPhoto():
				mediaGroup = append(mediaGroup, tgbotapi.NewInputMediaPhoto(tgbotapi.FileID(m.TGFileID)))
			case m.IsVideo():
				mediaGroup = append(mediaGroup, tgbotapi.NewInputMediaVideo(tgbotapi.FileID(m.TGFileID)))
			case m.IsDocument():
				mediaGroup = append(mediaGroup, tgbotapi.NewInputMediaDocument(tgbotapi.FileID(m.TGFileID)))
			}
		}
		return tgbotapi.NewMediaGroup(chatID, mediaGroup)
	}

	return nil
}
