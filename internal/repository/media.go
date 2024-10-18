package repository

import (
	"github.com/damianino/dcw-shop-bot/internal/repository/models"
	"gorm.io/gorm"
)

type media struct {
	db *gorm.DB
}

var _ Media = (*media)(nil)

func NewMediaRepo(db *gorm.DB) Media {
	return &media{db: db}
}

func (m media) Create(media *models.Media) error {
	if tx := m.db.Create(media); tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (m media) GetAllByMessageID(msgID int64) ([]models.Media, error) {
	var res []models.Media
	tx := m.db.Where(&models.Media{MessageID: msgID}).Find(&res)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return res, nil
}
