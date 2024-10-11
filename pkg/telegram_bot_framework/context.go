package telegram_bot_framework

import "context"

const (
	DialogDetailsCtxKey = "dialog_details"
)

type ContextDialogDetails struct {
	ChatID   int64
	UserID   int64
	Username string
}

func ContextWithDialogDetails(ctx context.Context, details ContextDialogDetails) context.Context {
	return context.WithValue(ctx, DialogDetailsCtxKey, details)
}

func GetContextDialogDetails(ctx context.Context) ContextDialogDetails {
	v := ctx.Value(DialogDetailsCtxKey)
	if v == nil {
		return ContextDialogDetails{}
	}
	return v.(ContextDialogDetails)
}
