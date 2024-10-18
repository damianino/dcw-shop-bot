package repository

import (
	"github.com/damianino/dcw-shop-bot/internal/repository/models"
	"gorm.io/gorm"
)

type messages struct {
	db *gorm.DB
}

var _ Messages = (*messages)(nil)

func NewMessagesRepo(db *gorm.DB) Messages {
	return &messages{db: db}
}

func (m messages) GetAllByDialogID(dialogID int64) ([]models.Message, error) {
	var res []models.Message
	tx := m.db.Where(&models.Message{DialogID: dialogID}).Find(&res)

	if tx.Error != nil {
		return nil, tx.Error
	}

	return res, nil
}

func (m messages) Create(message *models.Message) error {
	if tx := m.db.Create(message); tx.Error != nil {
		return tx.Error
	}

	return nil
}
