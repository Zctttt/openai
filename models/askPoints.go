package models

type AskPoint struct {
	Id     int64 `json:"id" gorm:"column:id"`
	UserId int64 `json:"user_id" gorm:"column:user_id"`
	Points int64 `json:"points" gorm:"column:points"`
}

func (m *AskPoint) TableName() string {
	return "ask_point"
}
