package handlers

import (
	"errors"

	"homework04/internal/database"
	"homework04/internal/middleware"
	"homework04/internal/models"
	"homework04/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreateCommentRequest struct {
	Content string `json:"content" binding:"required,min=1,max=1000"`
}

// POST /api/v1/posts/:id/comments
func CreateComment(c *gin.Context) {
	userID := middleware.CurrentUserID(c)

	postID, err := utils.ParseUintParam(c, "id")
	if err != nil {
		utils.Fail(c, err)
		return
	}

	var post models.Post
	if err := database.DB.Select("id").First(&post, postID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(c, utils.NotFound("文章不存在"))
			return
		}
		utils.Fail(c, utils.Internal(err))
		return
	}

	var req CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.BadRequest("请求参数不合法: "+err.Error()))
		return
	}

	comment := models.Comment{Content: req.Content, UserID: userID, PostID: postID}
	if err := database.DB.Create(&comment).Error; err != nil {
		utils.Fail(c, utils.Internal(err))
		return
	}

	var created models.Comment
	if err := database.DB.Preload("Author").First(&created, comment.ID).Error; err != nil {
		utils.Fail(c, utils.Internal(err))
		return
	}
	utils.Created(c, toCommentResponse(&created))
}

// GET /api/v1/posts/:id/comments
func ListComments(c *gin.Context) {
	postID, err := utils.ParseUintParam(c, "id")
	if err != nil {
		utils.Fail(c, err)
		return
	}

	var post models.Post
	if err := database.DB.Select("id").First(&post, postID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(c, utils.NotFound("文章不存在"))
			return
		}
		utils.Fail(c, utils.Internal(err))
		return
	}

	var comments []models.Comment
	err = database.DB.Preload("Author").
		Where("post_id = ?", postID).
		Order("id ASC").
		Find(&comments).Error
	if err != nil {
		utils.Fail(c, utils.Internal(err))
		return
	}

	items := make([]CommentResponse, 0, len(comments))
	for i := range comments {
		items = append(items, toCommentResponse(&comments[i]))
	}

	utils.OK(c, gin.H{"list": items, "total": len(items)})
}
