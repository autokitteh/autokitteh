package webtools

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/backend/db"
)

type terminalDB struct{ db.DB }

type Message struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	Addr string `json:"addr"`
	Text string `json:"text"`
}

func newDB(db db.DB) *terminalDB { return &terminalDB{db} }

func (db *terminalDB) Setup(ctx context.Context) error {
	return db.db(ctx).AutoMigrate(&Message{})
}

func (db *terminalDB) db(ctx context.Context) *gorm.DB {
	_, w := db.DB.GormDB()
	return w.WithContext(ctx)
}

func (db *terminalDB) GetMessages(ctx context.Context, addr string) ([]Message, error) {
	var msgs []Message

	if err := db.db(ctx).Where("addr = ?", addr).Find(&msgs).Error; err != nil {
		return nil, err
	}

	return msgs, nil
}

func (db *terminalDB) AddMessage(ctx context.Context, addr, text string) (uint, error) {
	if addr == "" {
		return 0, errors.New("addr is empty")
	}

	r := Message{Addr: addr, Text: text}

	if err := db.db(ctx).Create(&r).Error; err != nil {
		return 0, err
	}

	return r.ID, nil
}

// id == 0 -> delete all messages
func (db *terminalDB) DeleteMessage(ctx context.Context, addr string, id uint) error {
	q := db.db(ctx).Where("addr = ?", addr)

	if id != 0 {
		q = q.Where("id = ?", id)
	}

	if err := q.Delete(&Message{}).Error; err != nil {
		return err
	}

	return nil
}
