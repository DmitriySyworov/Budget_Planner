package model

import (
	"time"

	"gorm.io/gorm"
)

type BaseModel struct {
	CreatedAt time.Time      `gorm:"column:created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at"`
}
type DescriptionExpense struct {
	Category               string `gorm:"column:category"`
	Expense                string `gorm:"column:expense"`
	Description            string `gorm:"column:description"`
	DescriptionExpenseUUID string `gorm:"column:description_expense_uuid,primaryKey"`
	ExpenseUUID            string `gorm:"column:expense_uuid"`
}
type Expense struct {
	Health      string `json:"health,omitempty" gorm:"column:health"`
	Sport       string `json:"sport,omitempty"  gorm:"column:sport"`
	Supermarket string `json:"supermarket,omitempty"  gorm:"column:supermarket"`
	Restaurant  string `json:"restaurant,omitempty"  gorm:"column:restaurant"`
	Leisure     string `json:"leisure,omitempty"  gorm:"column:leisure"`
	Investments string `json:"investments,omitempty"  gorm:"column:investments"`
	Savings     string `json:"savings,omitempty"  gorm:"column:savings"`
	Other       string `json:"other,omitempty"   gorm:"column:other"`
	BudgetUUID  string `gorm:"column:budget_uuid"`
	ExpenseUUID string `gorm:"column:expense_uuid,primaryKey"`
}
type Budget struct {
	*BaseModel
	Amount      string    `gorm:"column:amount"`
	Start       time.Time `gorm:"column:start"`
	Finish      time.Time `gorm:"column:finish"`
	Description string    `gorm:"column:description"`
	BudgetUUID  string    `gorm:"column:budget_uuid"`
	UserUUID    string    `gorm:"column:user_uuid,primaryKey"`
}
type User struct {
	*BaseModel
	Name     string `gorm:"column:name"`
	Email    string `gorm:"column:email"`
	Password string `gorm:"column:password"`
	UserUUID string `gorm:"column:user_uuid,primaryKey"`
}
