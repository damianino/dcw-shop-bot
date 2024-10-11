package telegram_bot_framework

import (
	"context"
	"log"
)

type Action struct {
	name              string
	args              map[string]interface{}
	handler           Handler
	onSuccessRedirect *Page
	onFailureRedirect *Page
}

type Handler func(context.Context, DialogControls) error

func NewAction(name string,
	args map[string]interface{},
	handler Handler,
	onSuccessRedirect,
	onFailureRedirect *Page) *Action {
	return &Action{
		name,
		args,
		handler,
		onSuccessRedirect,
		onFailureRedirect,
	}
}

// Handle method returns a page to redirect to
func (a *Action) Handle(ctx context.Context, dialogControls DialogControls) *Page {
	err := a.handler(ctx, dialogControls)
	if err != nil {
		log.Printf("error handling action: %s\n", err.Error())

		return a.onFailureRedirect
	}

	return a.onSuccessRedirect
}
