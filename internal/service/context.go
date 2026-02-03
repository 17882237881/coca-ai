package service

import (
	"coca-ai/internal/llm"
	"coca-ai/internal/repository"
	"context"
	"fmt"
)

const (
	// 最大上下文消息数 (不包括系统提示和当前输入)
	MaxContextMessages = 20
	// 触发摘要压缩的阈值
	SummaryThreshold = 20
)

// ContextService 上下文构建服务
type ContextService struct {
	messageRepo  *repository.MessageRepository
	llmClient    llm.ChatClient
	systemPrompt string
}

// NewContextService 创建 ContextService 实例
func NewContextService(
	messageRepo *repository.MessageRepository,
	llmClient llm.ChatClient,
) *ContextService {
	return &ContextService{
		messageRepo: messageRepo,
		llmClient:   llmClient,
		systemPrompt: `你是 Coca AI，一个友好、专业的 AI 助手。
						你可以帮助用户解答问题、提供建议、进行对话。
						请使用简洁、清晰的语言回复用户。
						如果不确定答案，请诚实地表明。`,
	}
}

// BuildContext 构建 LLM 对话上下文
// 返回格式: [System Prompt, (可选)历史摘要, 最近消息..., 用户输入]
func (s *ContextService) BuildContext(ctx context.Context, sessionID int64, userInput string) ([]llm.Message, error) {
	// 1. 获取历史消息数量
	totalCount, err := s.messageRepo.CountBySessionID(ctx, sessionID)
	if err != nil {
		totalCount = 0
	}

	var messages []llm.Message

	// 2. 添加 System Prompt
	messages = append(messages, llm.Message{
		Role:    "system",
		Content: s.systemPrompt,
	})

	// 3. 如果历史消息较多，生成摘要
	if totalCount > SummaryThreshold {
		summary, err := s.generateSummary(ctx, sessionID)
		if err == nil && summary != "" {
			messages = append(messages, llm.Message{
				Role:    "system",
				Content: fmt.Sprintf("对话历史摘要：%s", summary),
			})
		}
	}

	// 4. 获取最近的消息
	recentMessages, err := s.messageRepo.FindRecentBySessionID(ctx, sessionID, MaxContextMessages)
	if err != nil {
		// 如果获取失败，继续执行（只有用户输入）
		recentMessages = nil
	}

	// 5. 将历史消息加入上下文
	for _, msg := range recentMessages {
		messages = append(messages, llm.Message{
			Role:    string(msg.Role),
			Content: msg.Content,
		})
	}

	// 6. 添加当前用户输入
	messages = append(messages, llm.Message{
		Role:    "user",
		Content: userInput,
	})

	return messages, nil
}

// generateSummary 生成历史对话摘要
func (s *ContextService) generateSummary(ctx context.Context, sessionID int64) (string, error) {
	// 获取所有历史消息（排除最近的）
	allMessages, err := s.messageRepo.FindBySessionID(ctx, sessionID)
	if err != nil {
		return "", err
	}

	// 只对前面的消息生成摘要，保留最近的 MaxContextMessages 条
	if len(allMessages) <= MaxContextMessages {
		return "", nil
	}

	messagesToSummarize := allMessages[:len(allMessages)-MaxContextMessages]

	// 转换为 LLM 消息格式
	llmMessages := make([]llm.Message, len(messagesToSummarize))
	for i, msg := range messagesToSummarize {
		llmMessages[i] = llm.Message{
			Role:    string(msg.Role),
			Content: msg.Content,
		}
	}

	// 调用 LLM 生成摘要
	return s.llmClient.Summarize(ctx, llmMessages)
}

// SetSystemPrompt 设置系统提示
func (s *ContextService) SetSystemPrompt(prompt string) {
	s.systemPrompt = prompt
}

// GetSystemPrompt 获取当前系统提示
func (s *ContextService) GetSystemPrompt() string {
	return s.systemPrompt
}
