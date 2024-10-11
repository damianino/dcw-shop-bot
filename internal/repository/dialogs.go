package repository

import (
	"github.com/damianino/dcw-shop-bot/internal/repository/models"
	"gorm.io/gorm"
)

type dialogs struct {
	db *gorm.DB
}

var _ Dialogs = (*dialogs)(nil)

func NewDialogsRepo(db *gorm.DB) Dialogs {
	return &dialogs{db: db}
}

func (d dialogs) Create(dialog *models.Dialog) error {
	if tx := d.db.Create(dialog); tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (d dialogs) GetAllUncompleted() ([]models.Dialog, error) {
	var res []models.Dialog
	tx := d.db.Model(&models.Dialog{}).Where(&models.Dialog{Completed: false}, "completed").Find(&res)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return res, nil
}

func (d dialogs) IsCompleted(dialog *models.Dialog) (bool, error) {
	tx := d.db.Where(dialog).Scan(dialog)
	if tx.Error != nil {
		return false, tx.Error
	}

	return dialog.Completed, nil
}

func (d dialogs) MarkCompleted(dialog *models.Dialog) error {
	tx := d.db.Model(&models.Dialog{}).Where(dialog).Update("completed", true).Scan(dialog)

	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (d dialogs) GetType(dialog *models.Dialog) (string, error) {
	tx := d.db.Where(dialog).Scan(dialog)
	if tx.Error != nil {
		return "", tx.Error
	}

	return dialog.DialogType, nil
}
