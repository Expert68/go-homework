package models

import (
	"time"
)

type BaseModel struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeleteAt  time.Time `gorm:"index" json:"-"`
}

type User struct {
	BaseModel
	Username string `gorm:"size:32;uniqueIndex;not null" json:"username"`
	Password string `gorm:"size:128;not null" json:"-"`
	Email    string `gorm:"size:128;uniqueIndex;not null" json:"email"`
}

type Post struct {
	BaseModel
	Title    string    `gorm:"size:255;not null" json:"title"`
	Content  string    `gorm:"type:text;not null" json:"content"`
	UserID   uint      `gorm:"index;not null" json:"user_id"`
	Author   User      `gorm:"foreignKey:UserID;not null" json:"author"`
	Comments []Comment `gorm:"foreignKey:PostID;not null" json:"comments,omitempty"`
}

type Comment struct {
	BaseModel
	Content string `gorm:"type:text;not null" json:"content"`
	UserID  uint   `gorm:"index;not null" json:"user_id"`
	PostID  uint   `gorm:"index;not null" json:"post_id"`
	Author  User   `gorm:"foreignKey:UserID;not null" json:"author"`
}
