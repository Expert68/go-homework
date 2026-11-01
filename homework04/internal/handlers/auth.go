package handlers

import (
	"errors"
	"homework04/internal/database"
	"homework04/internal/middleware"
	"homework04/internal/models"
	"homework04/internal/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=6,max=64"`
	Email    string `json:"email"    binding:"required,email"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.BadRequest("请求参数不合法："+err.Error()))
		return
	}

	var exists int64

	if err := database.DB.Model(&models.User{}).
		Where("username = ? OR email = ?", req.Username, req.Email).
		Count(&exists).Error; err != nil {
		utils.Fail(c, utils.Internal(err))
		return
	}

	if exists > 0 {
		utils.Fail(c, utils.Conflict("用户名或邮箱已注册"))
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	if err != nil {
		utils.Fail(c, utils.Internal(err))
		return
	}

	user := models.User{
		Username: req.Username,
		Password: string(hashed),
		Email:    req.Email,
	}
	if err := database.DB.Create(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			utils.Fail(c, utils.Conflict("用户名或邮箱已被注册"))
			return
		}
		utils.Fail(c, utils.Internal(err))
		return
	}
	utils.Created(c, toUserResponse(&user))
}

func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.BadRequest("请求参数不合法: "+err.Error()))
		return
	}

	var user models.User

	err := database.DB.Where("username = ?", req.Username).First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		utils.Fail(c, utils.Unauthorized("用户名或密码错误"))
		return
	}

	if err != nil {
		utils.Fail(c, utils.Internal(err))
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		utils.Fail(c, utils.Unauthorized("用户名或密码错误"))
		return
	}

	token, err := utils.GenerateToken(user.ID, user.Username)
	if err != nil {
		utils.Fail(c, utils.Internal(err))
		return
	}

	utils.OK(c,
		LoginResponse{
			Token:     token,
			ExpiresIn: int(utils.TOkenTTL().Seconds()),
			User:      toUserResponse(&user),
		})
}

// GET /api/v1/auth/me
func Me(c *gin.Context) {
	userID := middleware.CurrentUserID(c)
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(c, utils.NotFound("用户不存在"))
			return
		}
		utils.Fail(c, utils.Internal(err))
		return
	}
	utils.OK(c, toUserResponse(&user))
}
