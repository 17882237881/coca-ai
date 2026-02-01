package handler

import (
	"coca-ai/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ChatHandler 处理聊天相关的 HTTP 请求
type ChatHandler struct {
	chatSvc *service.ChatService
}

// NewChatHandler 创建 ChatHandler 实例
func NewChatHandler(chatSvc *service.ChatService) *ChatHandler {
	return &ChatHandler{chatSvc: chatSvc}
}

// ==================== Request/Response 结构体定义 ====================

// CreateSessionResp 创建会话响应
type CreateSessionResp struct {
	SessionID int64  `json:"session_id"`
	Title     string `json:"title"`
}

// SendMessageReq 发送消息请求
type SendMessageReq struct {
	Content string `json:"content" binding:"required"`
}

// SessionListItem 会话列表项
type SessionListItem struct {
	SessionID int64  `json:"session_id"`
	Title     string `json:"title"`
	UpdatedAt string `json:"updated_at"`
}

// MessageItem 消息列表项
type MessageItem struct {
	ID        int64  `json:"id"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

// ==================== 路由注册 ====================

// RegisterRoutes 注册聊天相关路由
func (h *ChatHandler) RegisterRoutes(server *gin.Engine, authMiddleware gin.HandlerFunc) {
	chatGroup := server.Group("/chat")
	chatGroup.Use(authMiddleware) // 所有聊天接口都需要登录
	{
		// 会话管理
		chatGroup.POST("/sessions", h.CreateSession)
		chatGroup.GET("/sessions", h.GetSessionList)
		chatGroup.DELETE("/sessions/:id", h.DeleteSession)

		// 消息管理
		chatGroup.GET("/sessions/:id/messages", h.GetMessages)
		chatGroup.POST("/sessions/:id/messages", h.SendMessage)
	}
}

// ==================== Handler 方法 ====================

// CreateSession 创建新会话
// POST /chat/sessions
func (h *ChatHandler) CreateSession(c *gin.Context) {
	userID := c.GetInt64("uid")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "Unauthorized"})
		return
	}

	session, err := h.chatSvc.CreateSession(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": CreateSessionResp{
			SessionID: session.ID,
			Title:     session.Title,
		},
	})
}

// GetSessionList 获取用户的会话列表
// GET /chat/sessions
func (h *ChatHandler) GetSessionList(c *gin.Context) {
	userID := c.GetInt64("uid")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "Unauthorized"})
		return
	}

	sessions, err := h.chatSvc.GetSessionList(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
		return
	}

	// 转换为响应格式
	items := make([]SessionListItem, len(sessions))
	for i, s := range sessions {
		items[i] = SessionListItem{
			SessionID: s.ID,
			Title:     s.Title,
			UpdatedAt: s.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": items,
	})
}

// GetMessages 获取会话的历史消息
// GET /chat/sessions/:id/messages
func (h *ChatHandler) GetMessages(c *gin.Context) {
	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "Invalid session ID"})
		return
	}

	messages, err := h.chatSvc.GetMessages(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
		return
	}

	// 转换为响应格式
	items := make([]MessageItem, len(messages))
	for i, m := range messages {
		items[i] = MessageItem{
			ID:        m.ID,
			Role:      string(m.Role),
			Content:   m.Content,
			CreatedAt: m.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": items,
	})
}

// SendMessage 发送消息并流式返回 AI 回复 (SSE)
// POST /chat/sessions/:id/messages
func (h *ChatHandler) SendMessage(c *gin.Context) {
	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "Invalid session ID"})
		return
	}

	var req SendMessageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "Invalid request: content is required"})
		return
	}

	// 设置 SSE 响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no") // 禁用 Nginx 缓冲

	// 调用 ChatService 发送消息，流式返回
	assistantMsg, err := h.chatSvc.SendMessage(c.Request.Context(), sessionID, req.Content, func(delta string) error {
		c.SSEvent("message", gin.H{"delta": delta})
		c.Writer.Flush()
		return nil
	})

	if err != nil {
		c.SSEvent("error", gin.H{"msg": err.Error()})
		c.Writer.Flush()
		return
	}

	// 发送完成事件
	c.SSEvent("done", gin.H{
		"message_id": assistantMsg.ID,
		"content":    assistantMsg.Content,
	})
	c.Writer.Flush()
}

// DeleteSession 删除会话
// DELETE /chat/sessions/:id
func (h *ChatHandler) DeleteSession(c *gin.Context) {
	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "Invalid session ID"})
		return
	}

	if err := h.chatSvc.DeleteSession(c.Request.Context(), sessionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "Session deleted",
	})
}
