package telegram_bot_framework

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

type Bot struct {
	bot            *tgbotapi.BotAPI
	dialogHandlers map[int64]*DialogHandler
	dialogTree     DialogTree
}

func NewBot(botAPI *tgbotapi.BotAPI, dialogTree DialogTree) *Bot {
	return &Bot{
		bot:            botAPI,
		dialogHandlers: make(map[int64]*DialogHandler),
		dialogTree:     dialogTree,
	}
}

func (b *Bot) StartBot(ctx context.Context) {
	b.consumeMessages(ctx)
}

func (b *Bot) consumeMessages(ctx context.Context) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.bot.GetUpdatesChan(u)

	var updatesIn chan tgbotapi.Update

	for update := range updates {
		if update.FromChat() == nil {
			continue
		}

		if _, ok := b.dialogHandlers[update.FromChat().ID]; !ok {
			b.dialogHandlers[update.FromChat().ID] = NewDialogHandler(
				b.dialogTree,
				make(chan tgbotapi.Update),
				make(chan tgbotapi.Chattable),
				make(chan struct{}, 1),
			)

			updatesIn = b.dialogHandlers[update.FromChat().ID].updatesIn

			ctx = ContextWithDialogDetails(ctx, ContextDialogDetails{
				ChatID:   update.FromChat().ID,
				UserID:   update.SentFrom().ID,
				Username: update.SentFrom().UserName,
			})

			go b.dialogHandlers[update.FromChat().ID].HandleDialog(ctx)
			go b.SendDialogMessages(ctx)
		}

		updatesIn <- update
	}
}

func (b *Bot) SendDialogMessages(ctx context.Context) {
	dd := GetContextDialogDetails(ctx)

	var deleteMessageConfig tgbotapi.DeleteMessageConfig
	var messageConfig tgbotapi.MessageConfig

	var deletePageMessageID int

	for chattable := range b.dialogHandlers[dd.ChatID].msgOut {
		switch chattable.(type) {
		case tgbotapi.MessageConfig:
			messageConfig = chattable.(tgbotapi.MessageConfig)
			messageConfig.ChatID = dd.ChatID
			sentMsg, err := b.bot.Send(messageConfig)
			if err != nil {
				log.Printf("failed to send message: %s\n", err.Error())
				continue
			}

			if deletePageMessageID != 0 {
				_, err = b.bot.Send(tgbotapi.NewDeleteMessage(dd.ChatID, deletePageMessageID))
				if err != nil {
					log.Printf("failed to send message: %s\n", err.Error())
				}
				deletePageMessageID = 0
			}

			select {
			case <-b.dialogHandlers[dd.ChatID].deleteFlag:
				deletePageMessageID = sentMsg.MessageID
			default:
			}
		case tgbotapi.DeleteMessageConfig:
			deleteMessageConfig = chattable.(tgbotapi.DeleteMessageConfig)
			deleteMessageConfig.ChatID = dd.ChatID
			_, err := b.bot.Send(deleteMessageConfig)
			if err != nil {
				log.Printf("failed to send message: %s\n", err.Error())
				continue
			}
		}
	}
}
