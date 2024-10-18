package models

import "time"

const (
	MediaTypeVideo    = "video"
	MediaTypePhoto    = "photo"
	MediaTypeDocument = "document"
)

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
	MessageID int64  `gorm:"primaryKey"`
	TGFileID  string `gorm:"column:tg_file_id"`
	Type      string
}

func (m *Media) IsPhoto() bool {
	return m.Type == MediaTypePhoto
}

func (m *Media) IsVideo() bool {
	return m.Type == MediaTypeVideo
}

func (m *Media) IsDocument() bool {
	return m.Type == MediaTypeDocument
}
