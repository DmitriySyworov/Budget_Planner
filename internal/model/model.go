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
	Health      string `json:"health,omitempty" gorm:"type:numeric(15, 2)"`
	Sport       string `json:"sport,omitempty" gorm:"type:numeric(15, 2)"`
	Supermarket string `json:"supermarket,omitempty" gorm:"type:numeric(15, 2)"`
	Restaurant  string `json:"restaurant,omitempty" gorm:"type:numeric(15, 2)"`
	Leisure     string `json:"leisure,omitempty" gorm:"type:numeric(15, 2)"`
	Investments string `json:"investments,omitempty" gorm:"type:numeric(15, 2)"`
	Savings     string `json:"savings,omitempty" gorm:"type:numeric(15, 2)"`
	Other       string `json:"other,omitempty"  gorm:"type:numeric(15, 2)"`
	BudgetUUID  string `gorm:"type:char(36)"`
	UserUUID    string `gorm:"type:char(36)"`
}
type Budget struct {
	*BaseModel
	Amount      string    `gorm:"type:numeric(15, 2)"`
	Start       time.Time `gorm:"type:date"`
	Finish      time.Time `gorm:"type:date"`
	Description string    `gorm:"type:text"`
	BudgetUUID  string    `gorm:"type:char(36)"`
	UserUUID    string    `gorm:"type:char(36),primaryKey"`
	Expenses    []Expense `gorm:"foreignKey:UserUUID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}
type User struct {
	*BaseModel
	Name     string `gorm:"column:name"`
	Email    string `gorm:"column:email"`
	Password string `gorm:"column:password"`
	UserUUID string `gorm:"column:user_uuid,primaryKey"`
}
