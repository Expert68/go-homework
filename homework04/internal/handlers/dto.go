package handlers

import (
	"homework04/internal/models"
	"time"
)

type UserBrief struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
}

type UserResponse struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type LoginResponse struct {
	Token     string       `json:"token"`
	ExpiresIn int          `json:"expires_in"`
	User      UserResponse `json:"user"`
}

type PostResponse struct {
	ID           uint      `json:"id"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	UserID       uint      `json:"user_id"`
	Author       UserBrief `json:"author"`
	CommentCount int64     `json:"comment_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CommentResponse struct {
	ID        uint      `json:"id"`
	Content   string    `json:"content"`
	UserID    uint      `json:"user_id"`
	PostID    uint      `json:"post_id"`
	Author    UserBrief `json:"author"`
	CreatedAt time.Time `json:"created_at"`
}

func toUserBrief(u models.User) UserBrief {
	return UserBrief{
		ID:       u.ID,
		Username: u.Username,
	}
}

func toUserResponse(u *models.User) UserResponse {
	return UserResponse{ID: u.ID, Username: u.Username, Email: u.Email, CreatedAt: u.CreatedAt}
}

func toPostResponse(p *models.Post, commentCount int64) PostResponse {
	return PostResponse{
		ID: p.ID, Title: p.Title, Content: p.Content, UserID: p.UserID,
		Author: toUserBrief(p.Author), CommentCount: commentCount,
		CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
	}
}

func toCommentResponse(cm *models.Comment) CommentResponse {
	return CommentResponse{
		ID: cm.ID, Content: cm.Content, UserID: cm.UserID, PostID: cm.PostID,
		Author: toUserBrief(cm.Author), CreatedAt: cm.CreatedAt,
	}
}
