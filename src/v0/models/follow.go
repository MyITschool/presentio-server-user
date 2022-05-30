package models

type Follow struct {
	ID     int64 `gorm:"primaryKey" json:"id"`
	FromID int64 `json:"fromId" binding:"required"`
	ToID   int64 `json:"toId" binding:"required"`
}
