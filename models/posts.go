package models

import (
	"time"
	"github.com/jinzhu/gorm"
	"go_bbs/services"
)

type Post struct {
	gorm.Model
	//ID        uint64 `gorm:"column:id"`
	User User `gorm:"ForeignKey:UserId;AssociationForeignKey:ID"`
	UserId int
	Replis []Reply
	Title string
	Content string
	ClickNumber int
	ReplyNumber int
	LastReplyAt time.Time
}

func (u *Post) TableName() string {
	return "posts"
}

func (u *Post) HotPosts(db  *services.DatabaseService, limitNum int) *[]Post {
	posts := []Post{}
	if err := db.DB.Find(u).Order("updated_at desc, last_reply_at desc").Limit(limitNum).Find(&posts).Error; err != nil {
		return nil
	}
	return &posts
}