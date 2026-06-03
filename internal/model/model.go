package model

import (
	"time"

	"gorm.io/gorm"
)

type BaseModel struct {
	CreatedAt time.Time      `gorm:"type:date"`
	UpdatedAt time.Time      `gorm:"type:date"`
	DeletedAt gorm.DeletedAt `gorm:"index;type:date"`
}
type Budget struct {
	*BaseModel
	Budget      float64   `gorm:"type:numeric(15, 2)"`
	Start       time.Time `gorm:"type:date"`
	Finish      time.Time `gorm:"type:date"`
	Description string    `gorm:"type:text"`
	BudgetUUID  string    `gorm:"type:char(36)"`
	UserUUID    string    `gorm:"type:char(36)"`
}
type User struct {
	*BaseModel
	Name     string `gorm:"type:varchar(64)"`
	Email    string `gorm:"type:varchar(256)"`
	Password string `gorm:"type:char(36)"`
	UserUUID string `gorm:"type:char(36)"`
}
