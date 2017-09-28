package models

import (
	"github.com/jinzhu/gorm"
)

type Reply struct {
	gorm.Model
	//ID        uint64 `gorm:"column:id"`
	//User User `gorm:"ForeignKey:UserId;AssociationForeignKey:ID"`
	User User
	UserId int
	Post Post
	PostId int
	PostTitle string
	Floor int
	Content string
}

func (u *Reply) TableName() string {
	return "replies"
}

