package models

type Expression struct {
	ID     uint    `gorm:"primaryKey"`
	Status string  `json:"status"`
	Result float64 `json:"result"`
}
