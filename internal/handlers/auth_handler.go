package handlers

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"github.com/jamesphm04/splose-clone-be/internal/services"
	"github.com/jamesphm04/splose-clone-be/internal/utils"
)

// AuthHandler handle authentication endpoints
type AuthHandler struct {
	userSvc  *services.UserService
	validate *validator.Validate
	log      *zap.Logger
}

func NewAuthHandler(userSvc *services.UserService, log *zap.Logger) *AuthHandler {
	return &AuthHandler{
		userSvc:  userSvc,
		validate: validator.New(),
		log:      log.Named("auth_handler"),
	}
}

// Register  POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var in services.RegisterInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.BadRequest(c, "invalid request body")
		return
	}

	if err := h.validate.Struct(in); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	user, err := h.userSvc.Register(c.Request.Context(), in)
	if err != nil {
		if errors.Is(err, services.ErrEmailTaken) {
			utils.Conflict(c, err.Error())
			return
		}

		h.log.Error("register failed", zap.Error(err))
		utils.InternalError(c)
		return
	}

	utils.Created(c, user)
}

// Login  POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var in services.LoginInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.BadRequest(c, "invalid request body")
		return
	}
	if err := h.validate.Struct(in); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	user, pair, err := h.userSvc.Login(c.Request.Context(), in)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			utils.Unauthorized(c, err.Error())
			return
		}
		h.log.Error("login failed", zap.Error(err))
		utils.InternalError(c)
		return
	}

	utils.OK(c, gin.H{
		"user": gin.H{
			"id":       user.ID,
			"email":    user.Email,
			"username": user.Username,
			"role":     user.Role,
		},
		"tokens": gin.H{
			"accessToken":  pair.AccessToken,
			"refreshToken": pair.RefreshToken,
		},
	})
}

// Refresh  POST /api/v1/auth/refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	var body struct {
		RefreshToken string `json:"refreshToken" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.BadRequest(c, "refreshToken is required")
		return
	}

	pair, err := h.userSvc.RefreshTokens(c.Request.Context(), body.RefreshToken)
	if err != nil {
		utils.Unauthorized(c, err.Error())
		return
	}

	utils.OK(c, pair)
}
