package service

import (
	"coca-ai/internal/domain"
	"coca-ai/internal/llm"
	"coca-ai/internal/mq"
	"coca-ai/internal/repository"
	"context"
	"time"
)

// ChatService 聊天核心业务服务
type ChatService struct {
	sessionRepo *repository.SessionRepository
	messageRepo *repository.MessageRepository
	llmClient   llm.ChatClient
	producer    *mq.Producer
	contextSvc  *ContextService
}

// NewChatService 创建 ChatService 实例
func NewChatService(
	sessionRepo *repository.SessionRepository,
	messageRepo *repository.MessageRepository,
	llmClient llm.ChatClient,
	producer *mq.Producer,
	contextSvc *ContextService,
) *ChatService {
	return &ChatService{
		sessionRepo: sessionRepo,
		messageRepo: messageRepo,
		llmClient:   llmClient,
		producer:    producer,
		contextSvc:  contextSvc,
	}
}

// StreamCallback 流式响应回调函数
type StreamCallback func(delta string) error

// CreateSession 创建新会话
func (s *ChatService) CreateSession(ctx context.Context, userID int64) (*domain.Session, error) {
	session := &domain.Session{
		UserID:    userID,
		Title:     "New Chat",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return s.sessionRepo.Create(ctx, session)
}

// GetSessionList 获取用户的会话列表
func (s *ChatService) GetSessionList(ctx context.Context, userID int64) ([]domain.Session, error) {
	return s.sessionRepo.FindByUserID(ctx, userID)
}

// GetMessages 获取会话的历史消息
func (s *ChatService) GetMessages(ctx context.Context, sessionID int64) ([]domain.Message, error) {
	return s.messageRepo.FindBySessionID(ctx, sessionID)
}

// DeleteSession 删除会话
func (s *ChatService) DeleteSession(ctx context.Context, sessionID int64) error {
	// 先删除消息
	if err := s.messageRepo.DeleteBySessionID(ctx, sessionID); err != nil {
		return err
	}
	// 再删除会话
	return s.sessionRepo.Delete(ctx, sessionID)
}

// SendMessage 发送消息并流式返回 AI 回复
func (s *ChatService) SendMessage(ctx context.Context, sessionID int64, content string, callback StreamCallback) (*domain.Message, error) {
	// 1. 创建用户消息
	userMsg := &domain.Message{
		SessionID: sessionID,
		Role:      domain.RoleUser,
		Content:   content,
		CreatedAt: time.Now(),
	}

	// 2. 写入 Redis 缓存 (热数据)
	if err := s.messageRepo.AppendToCache(ctx, userMsg); err != nil {
		// 缓存失败不阻塞流程，只记录日志
	}

	// 3. 发送到 Kafka (异步落库)
	if s.producer != nil {
		_ = s.producer.SendMessage(ctx, mq.DomainToEvent(userMsg))
	}

	// 4. 构建 LLM 上下文
	llmMessages, err := s.contextSvc.BuildContext(ctx, sessionID, content)
	if err != nil {
		return nil, err
	}

	// 5. 调用 LLM 流式接口
	var fullResponse string
	err = s.llmClient.StreamChat(ctx, llmMessages, func(delta string) error {
		fullResponse += delta
		if callback != nil {
			return callback(delta)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 6. 创建 AI 回复消息
	assistantMsg := &domain.Message{
		SessionID: sessionID,
		Role:      domain.RoleAssistant,
		Content:   fullResponse,
		CreatedAt: time.Now(),
	}

	// 7. 写入 Redis 缓存
	if err := s.messageRepo.AppendToCache(ctx, assistantMsg); err != nil {
		// 缓存失败不阻塞流程
	}

	// 8. 发送到 Kafka (异步落库)
	if s.producer != nil {
		_ = s.producer.SendMessage(ctx, mq.DomainToEvent(assistantMsg))
	}

	// 9. 更新会话的 updated_at
	_ = s.sessionRepo.TouchUpdatedAt(ctx, sessionID)

	// 10. 如果是第一条消息，自动生成会话标题
	count, _ := s.messageRepo.CountBySessionID(ctx, sessionID)
	if count <= 2 {
		// 用用户第一条消息作为标题（截取前 50 字符）
		title := content
		if len(title) > 50 {
			title = title[:50] + "..."
		}
		_ = s.sessionRepo.UpdateTitle(ctx, sessionID, title)
	}

	return assistantMsg, nil
}

// UpdateSessionTitle 更新会话标题
func (s *ChatService) UpdateSessionTitle(ctx context.Context, sessionID int64, title string) error {
	return s.sessionRepo.UpdateTitle(ctx, sessionID, title)
}
