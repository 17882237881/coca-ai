package handler

import (
	"github.com/gin-gonic/gin"
)

// UserHandler 定义用户相关的接口处理器
type UserHandler struct {
	// svc service.UserService // 暂时留空，后续注入
}

// NewUserHandler 初始化 UserHandler
func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

// RegisterRoutes 注册路由
func (h *UserHandler) RegisterRoutes(server *gin.Engine) {
	userGroup := server.Group("/users")
	{
		userGroup.POST("/signup", h.SignUp)
		userGroup.POST("/login", h.Login)
	}
}

// SignUpReq 注册请求结构体
type SignUpReq struct {
	Email           string `json:"email" binding:"required,email"`
	Password        string `json:"password" binding:"required,min=8"`
	ConfirmPassword string `json:"confirmPassword" binding:"required,eqfield=Password"`
}

// SignUp 用户注册接口定义
func (h *UserHandler) SignUp(ctx *gin.Context) {
	// TODO: 等待 Service 层完成后，在最后一步实现
}

// LoginReq 登录请求结构体
type LoginReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Login 用户登录接口定义
func (h *UserHandler) Login(ctx *gin.Context) {
	// TODO: 等待 Service 层完成后，在最后一步实现
}
