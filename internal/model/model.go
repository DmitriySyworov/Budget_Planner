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
type Expense struct {
	Health      float64 `gorm:"type:numeric(15, 2)"`
	Sport       float64 `gorm:"type:numeric(15, 2)"`
	Supermarket float64 `gorm:"type:numeric(15, 2)"`
	Restaurant  float64 `gorm:"type:numeric(15, 2)"`
	Leisure     float64 `gorm:"type:numeric(15, 2)"`
	Investments float64 `gorm:"type:numeric(15, 2)"`
	BudgetUUID  string  `gorm:"type:char(36)"`
	UserUUID    string  `gorm:"type:char(36)"`
}
type Budget struct {
	*BaseModel
	Amount      float64   `gorm:"type:numeric(15, 2)"`
	Start       time.Time `gorm:"type:date"`
	Finish      time.Time `gorm:"type:date"`
	Description string    `gorm:"type:text"`
	BudgetUUID  string    `gorm:"type:char(36)"`
	UserUUID    string    `gorm:"type:char(36),primaryKey"`
	Expenses    []Expense `gorm:"foreignKey:UserUUID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}
type User struct {
	*BaseModel
	Name     string   `gorm:"type:varchar(64)"`
	Email    string   `gorm:"type:varchar(256)"`
	Password string   `gorm:"type:char(36)"`
	UserUUID string   `gorm:"type:char(36)"`
	Budgets  []Budget `gorm:"foreignKey:UserUUID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}
