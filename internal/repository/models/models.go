package models

import "time"

type Dialog struct {
	ID               int64  `gorm:"primaryKey"`
	TGUserID         int64  `gorm:"column:tg_user_id"`
	TGChatID         int64  `gorm:"column:tg_chat_id"`
	TGClientUsername string `gorm:"column:tg_client_username"`
	Completed        bool
	DialogType       string `gorm:"column:type"`
	CreatedAt        time.Time
	CompletedAt      *time.Time `gorm:"index"`
}

type Message struct {
	ID             int64 `gorm:"primaryKey"`
	DialogID       int64
	SentByCustomer bool
	Text           *string
	CreatedAt      time.Time `gorm:"index"`
}

type Media struct {
	MessageID int64 `gorm:"primaryKey"`
	TGFileID  int64 `gorm:"column:tg_file_id"`
}
