package handlers

import (
	"strconv"

	"package-base/internal/core/interfaces"
	"package-base/internal/core/utils"
	"package-base/internal/dto"
	"package-base/internal/middleware"
	"package-base/internal/models"
	serviceInterfaces "package-base/internal/services/interfaces"

	"github.com/gofiber/fiber/v2"
)

// UserHandler 用户处理器
type UserHandler struct {
	userService serviceInterfaces.UserService
	validator   *utils.Validator
	logger      interfaces.Logger
}

// NewUserHandler 创建用户处理器
func NewUserHandler(
	userService serviceInterfaces.UserService,
	logger interfaces.Logger,
) *UserHandler {
	return &UserHandler{
		userService: userService,
		validator:   utils.NewValidator(),
		logger:      logger,
	}
}

// GetProfile 获取用户资料
// @Summary 获取用户资料
// @Description 获取当前用户的详细资料信息
// @Tags 用户
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dto.APIResponse{data=dto.UserInfo}
// @Failure 401 {object} dto.APIResponse
// @Failure 404 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/user/profile [get]
func (h *UserHandler) GetProfile(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	userInfo, err := h.userService.GetProfile(c.Context(), userID)
	if err != nil {
		h.logger.Error("failed to get user profile",
			interfaces.ErrorField(err),
			interfaces.String("user_id", userID),
		)

		statusCode := fiber.StatusInternalServerError
		message := "failed to get user profile"

		if err.Error() == "user not found" {
			statusCode = fiber.StatusNotFound
			message = err.Error()
		}

		return c.Status(statusCode).JSON(dto.Error(statusCode, message))
	}

	return c.JSON(dto.Success(userInfo))
}

// UpdateProfile 更新用户资料
// @Summary 更新用户资料
// @Description 更新当前用户的资料信息
// @Tags 用户
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.UpdateProfileRequest true "用户资料"
// @Success 200 {object} dto.APIResponse{data=dto.UserInfo}
// @Failure 400 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Failure 404 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/user/profile [put]
func (h *UserHandler) UpdateProfile(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	var req dto.UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Warn("failed to parse update profile request",
			interfaces.ErrorField(err),
			interfaces.String("user_id", userID),
		)
		return c.Status(fiber.StatusBadRequest).JSON(dto.Error(400, "invalid request body"))
	}

	// 验证请求参数
	if err := h.validator.ValidateStruct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.Error(400, err.Error()))
	}

	// 调用服务层更新资料
	userInfo, err := h.userService.UpdateProfile(c.Context(), userID, &req)
	if err != nil {
		h.logger.Error("failed to update user profile",
			interfaces.ErrorField(err),
			interfaces.String("user_id", userID),
		)

		statusCode := fiber.StatusInternalServerError
		message := "failed to update user profile"

		if err.Error() == "user not found" {
			statusCode = fiber.StatusNotFound
			message = err.Error()
		}

		return c.Status(statusCode).JSON(dto.Error(statusCode, message))
	}

	h.logger.Info("user profile updated successfully",
		interfaces.String("user_id", userID),
	)

	return c.JSON(dto.SuccessWithMessage("profile updated successfully", userInfo))
}

// GetUserByID 根据ID获取用户信息（管理员功能）
// @Summary 获取指定用户信息
// @Description 管理员根据用户ID获取用户信息
// @Tags 用户管理
// @Security BearerAuth
// @Param userId path string true "用户ID"
// @Produce json
// @Success 200 {object} dto.APIResponse{data=dto.UserInfo}
// @Failure 400 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Failure 403 {object} dto.APIResponse
// @Failure 404 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/admin/users/{userId} [get]
func (h *UserHandler) GetUserByID(c *fiber.Ctx) error {
	targetUserID := c.Params("userId")
	if targetUserID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.Error(400, "user ID is required"))
	}

	user, err := h.userService.GetByID(c.Context(), targetUserID)
	if err != nil {
		h.logger.Error("failed to get user by ID",
			interfaces.ErrorField(err),
			interfaces.String("target_user_id", targetUserID),
			interfaces.String("admin_user_id", middleware.GetUserID(c)),
		)

		statusCode := fiber.StatusInternalServerError
		message := "failed to get user"

		if err.Error() == "user not found" {
			statusCode = fiber.StatusNotFound
			message = err.Error()
		}

		return c.Status(statusCode).JSON(dto.Error(statusCode, message))
	}

	userInfo := h.userService.ConvertToUserInfo(user)
	return c.JSON(dto.Success(userInfo))
}

// GetUserList 获取用户列表（管理员功能）
// @Summary 获取用户列表
// @Description 管理员获取分页用户列表
// @Tags 用户管理
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Produce json
// @Success 200 {object} dto.APIResponse{data=dto.PaginationResponse}
// @Failure 400 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Failure 403 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/admin/users [get]
func (h *UserHandler) GetUserList(c *fiber.Ctx) error {
	// 获取分页参数
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		return c.Status(fiber.StatusBadRequest).JSON(dto.Error(400, "invalid page parameter"))
	}

	pageSize, err := strconv.Atoi(c.Query("page_size", "20"))
	if err != nil || pageSize < 1 || pageSize > 100 {
		return c.Status(fiber.StatusBadRequest).JSON(dto.Error(400, "invalid page_size parameter (1-100)"))
	}

	// 调用服务层获取用户列表
	result, err := h.userService.GetUserList(c.Context(), page, pageSize)
	if err != nil {
		h.logger.Error("failed to get user list",
			interfaces.ErrorField(err),
			interfaces.String("admin_user_id", middleware.GetUserID(c)),
			interfaces.Int("page", page),
			interfaces.Int("page_size", pageSize),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(dto.Error(500, "failed to get user list"))
	}

	return c.JSON(dto.Success(result))
}

// UpdateUserStatus 更新用户状态（管理员功能）
// @Summary 更新用户状态
// @Description 管理员更新指定用户的状态
// @Tags 用户管理
// @Security BearerAuth
// @Param userId path string true "用户ID"
// @Param request body object{status=int} true "状态信息"
// @Produce json
// @Success 200 {object} dto.APIResponse
// @Failure 400 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Failure 403 {object} dto.APIResponse
// @Failure 404 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/admin/users/{userId}/status [put]
func (h *UserHandler) UpdateUserStatus(c *fiber.Ctx) error {
	targetUserID := c.Params("userId")
	if targetUserID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.Error(400, "user ID is required"))
	}

	var req struct {
		Status int `json:"status" validate:"required,min=0,max=3"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.Error(400, "invalid request body"))
	}

	// 验证请求参数
	if err := h.validator.ValidateStruct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.Error(400, err.Error()))
	}

	// 调用服务层更新状态
	status := models.UserStatus(req.Status)
	if err := h.userService.UpdateStatus(c.Context(), targetUserID, status); err != nil {
		h.logger.Error("failed to update user status",
			interfaces.ErrorField(err),
			interfaces.String("target_user_id", targetUserID),
			interfaces.String("admin_user_id", middleware.GetUserID(c)),
			interfaces.Int("new_status", req.Status),
		)

		statusCode := fiber.StatusInternalServerError
		message := "failed to update user status"

		if err.Error() == "user not found" {
			statusCode = fiber.StatusNotFound
			message = err.Error()
		}

		return c.Status(statusCode).JSON(dto.Error(statusCode, message))
	}

	h.logger.Info("user status updated successfully",
		interfaces.String("target_user_id", targetUserID),
		interfaces.String("admin_user_id", middleware.GetUserID(c)),
		interfaces.Int("new_status", req.Status),
	)

	return c.JSON(dto.Success("user status updated successfully"))
}

// DeleteUser 删除用户（管理员功能）
// @Summary 删除用户
// @Description 管理员删除指定用户（软删除）
// @Tags 用户管理
// @Security BearerAuth
// @Param userId path string true "用户ID"
// @Produce json
// @Success 200 {object} dto.APIResponse
// @Failure 400 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Failure 403 {object} dto.APIResponse
// @Failure 404 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/admin/users/{userId} [delete]
func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	targetUserID := c.Params("userId")
	if targetUserID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.Error(400, "user ID is required"))
	}

	// 防止管理员删除自己
	adminUserID := middleware.GetUserID(c)
	if targetUserID == adminUserID {
		return c.Status(fiber.StatusBadRequest).JSON(dto.Error(400, "cannot delete yourself"))
	}

	// 调用服务层删除用户
	if err := h.userService.DeleteUser(c.Context(), targetUserID); err != nil {
		h.logger.Error("failed to delete user",
			interfaces.ErrorField(err),
			interfaces.String("target_user_id", targetUserID),
			interfaces.String("admin_user_id", adminUserID),
		)

		statusCode := fiber.StatusInternalServerError
		message := "failed to delete user"

		if err.Error() == "user not found" {
			statusCode = fiber.StatusNotFound
			message = err.Error()
		}

		return c.Status(statusCode).JSON(dto.Error(statusCode, message))
	}

	h.logger.Info("user deleted successfully",
		interfaces.String("target_user_id", targetUserID),
		interfaces.String("admin_user_id", adminUserID),
	)

	return c.JSON(dto.Success("user deleted successfully"))
}

// SearchUsers 搜索用户（管理员功能）
// @Summary 搜索用户
// @Description 管理员根据关键词搜索用户
// @Tags 用户管理
// @Security BearerAuth
// @Param q query string true "搜索关键词"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Produce json
// @Success 200 {object} dto.APIResponse{data=dto.PaginationResponse}
// @Failure 400 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Failure 403 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/admin/users/search [get]
func (h *UserHandler) SearchUsers(c *fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.Error(400, "search query is required"))
	}

	// 获取分页参数
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		return c.Status(fiber.StatusBadRequest).JSON(dto.Error(400, "invalid page parameter"))
	}

	pageSize, err := strconv.Atoi(c.Query("page_size", "20"))
	if err != nil || pageSize < 1 || pageSize > 100 {
		return c.Status(fiber.StatusBadRequest).JSON(dto.Error(400, "invalid page_size parameter (1-100)"))
	}

	// 这里可以扩展为实际的搜索功能
	// 目前返回普通的用户列表
	result, err := h.userService.GetUserList(c.Context(), page, pageSize)
	if err != nil {
		h.logger.Error("failed to search users",
			interfaces.ErrorField(err),
			interfaces.String("admin_user_id", middleware.GetUserID(c)),
			interfaces.String("query", query),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(dto.Error(500, "failed to search users"))
	}

	return c.JSON(dto.Success(result))
}
