package handlers

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"github.com/jamesphm04/splose-clone-be/internal/middleware"
	"github.com/jamesphm04/splose-clone-be/internal/repositories"
	"github.com/jamesphm04/splose-clone-be/internal/services"
	"github.com/jamesphm04/splose-clone-be/internal/utils"
)

// UserHandler handles user management endpoints.
type UserHandler struct {
	userSvc  *services.UserService
	validate *validator.Validate
	log      *zap.Logger
}

func NewUserHandler(userSvc *services.UserService, log *zap.Logger) *UserHandler {
	return &UserHandler{userSvc: userSvc, validate: validator.New(), log: log.Named("user_handler")}
}

// GetMe  GET /api/v1/users/me
func (h *UserHandler) GetMe(c *gin.Context) {
	user, err := h.userSvc.GetByID(c.Request.Context(), middleware.GetUserID(c))
	if err != nil {
		utils.NotFound(c, "user")
		return
	}
	utils.OK(c, user)
}

// List  GET /api/v1/users  (admin only)
func (h *UserHandler) List(c *gin.Context) {
	page, pageSize, offset := utils.Pagination(c)
	users, total, err := h.userSvc.List(c.Request.Context(), offset, pageSize)
	if err != nil {
		h.log.Error("list users failed", zap.Error(err))
		utils.InternalError(c)
		return
	}
	utils.OKList(c, users, utils.BuildMeta(page, pageSize, total))
}

// Update  PATCH /api/v1/users/:id
func (h *UserHandler) Update(c *gin.Context) {
	id := c.Param("id")
	callerID := middleware.GetUserID(c)
	callerRole := middleware.GetRole(c)

	if id != callerID && callerRole != "admin" {
		utils.Forbidden(c)
		return
	}

	var in services.UpdateUserInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.BadRequest(c, "invalid request body")
		return
	}
	if err := h.validate.Struct(in); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	user, err := h.userSvc.Update(c.Request.Context(), id, in)
	if err != nil {
		switch {
		case errors.Is(err, repositories.ErrNotFound):
			utils.NotFound(c, "user")
		case errors.Is(err, services.ErrEmailTaken):
			utils.Conflict(c, "email already taken")
		default:
			h.log.Error("update user failed", zap.String("userID", id), zap.Error(err))
			utils.InternalError(c)
		}
		return
	}

	utils.OK(c, user)
}

// Delete  DELETE /api/v1/users/:id
func (h *UserHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	callerID := middleware.GetUserID(c)
	callerRole := middleware.GetRole(c)

	if id != callerID && callerRole != "admin" {
		utils.Forbidden(c)
		return
	}

	if err := h.userSvc.SoftDelete(c.Request.Context(), id); err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			utils.NotFound(c, "user")
			return
		}
		h.log.Error("delete user failed", zap.String("userID", id), zap.Error(err))
		utils.InternalError(c)
		return
	}

	utils.OK(c, gin.H{"message": "user deleted"})
}
