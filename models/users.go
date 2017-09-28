package models

import (
	"github.com/jinzhu/gorm"
)

type User struct {
	gorm.Model
	//ID        uint64 `gorm:"column:id"`
	Name      string
	Email     string
	Password  string
	Posts []Post
}

func (u *User) TableName() string {
	return "users"
}


