package telegram_bot_framework

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type DialogHandler struct {
	dialogTree  *DialogTree
	currentPage *Page
	updatesIn   chan tgbotapi.Update
	msgOut      chan tgbotapi.Chattable
	deleteFlag  chan struct{}
}

func NewDialogHandler(dialogTree DialogTree, msgIn chan tgbotapi.Update, msgOut chan tgbotapi.Chattable, deleteFlag chan struct{}) *DialogHandler {
	return &DialogHandler{
		dialogTree:  &dialogTree,
		currentPage: dialogTree.root,
		updatesIn:   msgIn,
		msgOut:      msgOut,
		deleteFlag:  deleteFlag,
	}
}

func (dh *DialogHandler) HandleDialog(ctx context.Context) {
	redirect := dh.currentPage
	waitForOk := false

	dc := DialogControls{
		UpdatesIn: dh.updatesIn,
		MsgOut:    dh.msgOut,
	}

	for update := range dh.updatesIn {
		switch {
		// command
		case update.Message != nil && update.Message.From != nil && update.Message.IsCommand():

		// message
		case update.Message != nil && update.Message.From != nil:

		// callback
		case update.CallbackQuery != nil:
			if !waitForOk {
				redirect, waitForOk = dh.handleButton(ctx, update, dc, redirect)
			}
		}

		if waitForOk {
			waitForOk = false
			continue
		}

		dh.currentPage = redirect
		dh.deleteFlag <- struct{}{}
		dh.msgOut <- redirect.BuildMessage()
	}
}

func (dh *DialogHandler) handleButton(ctx context.Context, update tgbotapi.Update, dc DialogControls, redirect *Page) (*Page, bool) {
	var btn button
	var waitForOk bool

	if dh.currentPage.backPage != nil && update.CallbackData() == dh.currentPage.backPage.id {
		return dh.currentPage.backPage, false
	}

	for _, b := range dh.currentPage.buttons {
		if update.CallbackQuery.Data == b.callbackData {
			btn = b
			break
		}
	}

	if btn.action != nil {
		redirect = btn.action.Handle(ctx, dc)
	}

	if btn.page != nil {
		redirect = btn.page
	}

	if btn.defaultResponse != nil {
		waitForOk = true
		for i, msg := range btn.defaultResponse {
			if i == len(btn.defaultResponse)-1 {
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Ok", "ok"),
					),
				)
			}
			redirect = dh.currentPage
			dh.msgOut <- msg
		}

	}

	return redirect, waitForOk
}
