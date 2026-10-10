package handlers

import (
	"errors"
	"strings"

	"homework04/internal/database"
	"homework04/internal/middleware"
	"homework04/internal/models"
	"homework04/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreatePostRequest struct {
	Title   string `json:"title"   binding:"required,min=1,max=200"`
	Content string `json:"content" binding:"required,min=1"`
}

type UpdatePostRequest struct {
	Title   *string `json:"title"   binding:"omitempty,min=1,max=200"`
	Content *string `json:"content" binding:"omitempty,min=1"`
}

func ListPosts(c *gin.Context) {
	page := utils.QueryInt(c, "page", 1)
	size := utils.QueryInt(c, "size", 10)
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 50 {
		size = 10
	}

	var total int64
	if err := buildPostQuery(c).Count(&total).Error; err != nil {
		utils.Fail(c, utils.Internal(err))
		return
	}

	var posts []models.Post
	err := buildPostQuery(c).
		Preload("Author").
		Order("id DESC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&posts).Error
	if err != nil {
		utils.Fail(c, utils.Internal(err))
		return
	}

	ids := make([]uint, 0, len(posts))
	for _, p := range posts {
		ids = append(ids, p.ID)
	}
	counts, err := countCommentsByPostIDs(ids)
	if err != nil {
		utils.Fail(c, utils.Internal(err))
		return
	}

	items := make([]PostResponse, 0, len(posts))
	for i := range posts {
		items = append(items, toPostResponse(&posts[i], counts[posts[i].ID]))
	}

	utils.OK(c, gin.H{
		"list":  items,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

func buildPostQuery(c *gin.Context) *gorm.DB {
	q := database.DB.Model(&models.Post{})
	if kw := strings.TrimSpace(c.Query("keyword")); kw != "" {
		q = q.Where("title LIKE ?", "%"+kw+"%")
	}
	return q
}

func GetPost(c *gin.Context) {
	id, err := utils.ParseUintParam(c, "id")
	if err != nil {
		utils.Fail(c, err)
		return
	}

	var post models.Post
	if err := database.DB.Preload("Author").First(&post, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(c, utils.NotFound("文章不存在"))
			return
		}
		utils.Fail(c, utils.Internal(err))
		return
	}

	counts, _ := countCommentsByPostIDs([]uint{post.ID})
	utils.OK(c, toPostResponse(&post, counts[post.ID]))
}

func CreatePost(c *gin.Context) {
	userID := middleware.CurrentUserID(c) // ← 认证中间件已经帮我们准备好了

	var req CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.BadRequest("请求参数不合法: "+err.Error()))
		return
	}

	post := models.Post{Title: req.Title, Content: req.Content, UserID: userID}
	if err := database.DB.Create(&post).Error; err != nil {
		utils.Fail(c, utils.Internal(err))
		return
	}

	var created models.Post
	if err := database.DB.Preload("Author").First(&created, post.ID).Error; err != nil {
		utils.Fail(c, utils.Internal(err))
		return
	}
	utils.Created(c, toPostResponse(&created, 0))
}

func UpdatePost(c *gin.Context) {
	userID := middleware.CurrentUserID(c)

	id, err := utils.ParseUintParam(c, "id")
	if err != nil {
		utils.Fail(c, err)
		return
	}

	var post models.Post
	if err := database.DB.First(&post, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(c, utils.NotFound("文章不存在"))
			return
		}
		utils.Fail(c, utils.Internal(err))
		return
	}

	if post.UserID != userID {
		utils.Fail(c, utils.Forbidden("只能修改自己的文章"))
		return
	}

	var req UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.BadRequest("请求参数不合法: "+err.Error()))
		return
	}

	updates := map[string]any{}
	if req.Title != nil {
		updates["title"] = *req.Title // * 是"解引用"，取出指针指向的值
	}
	if req.Content != nil {
		updates["content"] = *req.Content
	}
	if len(updates) == 0 {
		utils.Fail(c, utils.BadRequest("没有需要更新的字段"))
		return
	}
	if err := database.DB.Model(&post).Updates(updates).Error; err != nil {
		utils.Fail(c, utils.Internal(err))
		return
	}

	var updated models.Post
	if err := database.DB.Preload("Author").First(&updated, post.ID).Error; err != nil {
		utils.Fail(c, utils.Internal(err))
		return
	}
	counts, _ := countCommentsByPostIDs([]uint{updated.ID})
	utils.OK(c, toPostResponse(&updated, counts[updated.ID]))
}

func DeletePost(c *gin.Context) {
	userID := middleware.CurrentUserID(c)

	id, err := utils.ParseUintParam(c, "id")
	if err != nil {
		utils.Fail(c, err)
		return
	}

	var post models.Post
	if err := database.DB.First(&post, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(c, utils.NotFound("文章不存在"))
			return
		}
		utils.Fail(c, utils.Internal(err))
		return
	}
	if post.UserID != userID {
		utils.Fail(c, utils.Forbidden("只能删除自己的文章"))
		return
	}

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("post_id = ?", post.ID).Delete(&models.Comment{}).Error; err != nil {
			return err
		}
		return tx.Delete(&post).Error
	})
	if err != nil {
		utils.Fail(c, utils.Internal(err))
		return
	}

	utils.OK(c, gin.H{"deleted": true, "id": post.ID})
}

type commentCountRow struct {
	PostID uint  `gorm:"column:post_id"`
	Total  int64 `gorm:"column:total"`
}

func countCommentsByPostIDs(ids []uint) (map[uint]int64, error) {
	result := make(map[uint]int64, len(ids))
	if len(ids) == 0 {
		return result, nil
	}
	var rows []commentCountRow
	err := database.DB.Model(&models.Comment{}).
		Select("post_id, COUNT(*) AS total").
		Where("post_id IN ?", ids).
		Group("post_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		result[r.PostID] = r.Total
	}
	return result, nil
}
