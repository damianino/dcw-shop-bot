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
	tx := m.db.Where(&models.Media{MessageID: msgID})
	if tx.Error != nil {
		return nil, tx.Error
	}

	rows, err := tx.Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]models.Media, tx.RowsAffected)
	for rows.Next() {
		if err := rows.Scan(&res); err != nil {
			return nil, err
		}
	}

	return res, nil
}
