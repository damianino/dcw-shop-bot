package repository

import (
	"github.com/damianino/dcw-shop-bot/internal/repository/models"
	"gorm.io/gorm"
)

type Repository struct {
	Dialogs
	Messages
	Media
}

type Dialogs interface {
	Create(dialog *models.Dialog) error
	GetAllUncompleted() ([]models.Dialog, error)
	IsCompleted(dialog *models.Dialog) (bool, error)
	MarkCompleted(dialog *models.Dialog) error
	GetType(dialog *models.Dialog) (string, error)
}

type Messages interface {
	Create(message *models.Message) error
	GetAllByDialogID(dialogID int64) ([]models.Message, error)
}

type Media interface {
	Create(media *models.Media) error
	GetAllByMessageID(msgID int64) ([]models.Media, error)
}

type Common interface {
}

func NewRepo(db *gorm.DB) *Repository {
	return &Repository{
		Dialogs:  NewDialogsRepo(db),
		Messages: NewMessagesRepo(db),
		Media:    NewMediaRepo(db),
	}
}
