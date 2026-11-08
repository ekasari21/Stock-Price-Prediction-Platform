package model

import (
	"time"
)

type Stock struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Symbol      string    `gorm:"type:varchar(20);uniqueIndex;not null" json:"symbol"`
	Name        string    `gorm:"type:varchar(200);not null" json:"name"`
	Sector      string    `gorm:"type:varchar(100)" json:"sector"`
	Exchange    string    `gorm:"type:varchar(50)" json:"exchange"`
	Currency    string    `gorm:"type:varchar(10);default:'USD'" json:"currency"`
	LastPrice   float64   `gorm:"type:decimal(12,2)" json:"last_price"`
	PriceChange float64   `gorm:"type:decimal(12,2)" json:"price_change"`
	PctChange   float64   `gorm:"type:decimal(8,2)" json:"pct_change"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
