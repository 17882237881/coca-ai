package llm

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
)

// QwenClient 通义千问客户端 (通过 Eino 框架)
type QwenClient struct {
	chatModel *openai.ChatModel
	model     string
}

// QwenConfig 通义千问配置
type QwenConfig struct {
	APIKey  string // 通义千问 API Key
	BaseURL string // API 端点 (通义千问兼容 OpenAI 格式)
	Model   string // 模型名称，如 "qwen-plus", "qwen-turbo"
}

// NewQwenClient 创建通义千问客户端
func NewQwenClient(cfg *QwenConfig) (*QwenClient, error) {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
	}
	if cfg.Model == "" {
		cfg.Model = "qwen-plus"
	}

	chatModel, err := openai.NewChatModel(context.Background(), &openai.ChatModelConfig{
		BaseURL: cfg.BaseURL,
		APIKey:  cfg.APIKey,
		Model:   cfg.Model,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Qwen client: %w", err)
	}

	return &QwenClient{
		chatModel: chatModel,
		model:     cfg.Model,
	}, nil
}

// Chat 普通对话 (非流式)
func (c *QwenClient) Chat(ctx context.Context, messages []Message) (string, error) {
	einoMessages := c.convertMessages(messages)

	resp, err := c.chatModel.Generate(ctx, einoMessages)
	if err != nil {
		return "", fmt.Errorf("chat failed: %w", err)
	}

	if resp == nil || resp.Content == "" {
		return "", fmt.Errorf("empty response from LLM")
	}

	return resp.Content, nil
}

// StreamChat 流式对话
func (c *QwenClient) StreamChat(ctx context.Context, messages []Message, callback StreamCallback) error {
	einoMessages := c.convertMessages(messages)

	stream, err := c.chatModel.Stream(ctx, einoMessages)
	if err != nil {
		return fmt.Errorf("stream chat failed: %w", err)
	}
	defer stream.Close()

	for {
		chunk, err := stream.Recv() 
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("stream recv failed: %w", err)
		}

		if chunk != nil && chunk.Content != "" {
			if err := callback(chunk.Content); err != nil {
				return err
			}
		}
	}

	return nil
}

// Summarize 生成摘要
func (c *QwenClient) Summarize(ctx context.Context, messages []Message) (string, error) {
	// 构建摘要请求
	var content strings.Builder
	content.WriteString("请将以下对话内容总结为简短的摘要（不超过200字）：\n\n")

	for _, msg := range messages {
		switch msg.Role {
		case "user":
			content.WriteString(fmt.Sprintf("用户: %s\n", msg.Content))
		case "assistant":
			content.WriteString(fmt.Sprintf("助手: %s\n", msg.Content))
		}
	}

	summaryMessages := []Message{
		{Role: "system", Content: "你是一个专业的摘要助手，请用简洁的语言总结对话的主要内容。"},
		{Role: "user", Content: content.String()},
	}

	return c.Chat(ctx, summaryMessages)
}

// convertMessages 将内部 Message 转换为 Eino schema.Message
func (c *QwenClient) convertMessages(messages []Message) []*schema.Message {
	result := make([]*schema.Message, len(messages))
	for i, msg := range messages {
		var role schema.RoleType
		switch msg.Role {
		case "user":
			role = schema.User
		case "assistant":
			role = schema.Assistant
		case "system":
			role = schema.System
		default:
			role = schema.User
		}
		result[i] = &schema.Message{
			Role:    role,
			Content: msg.Content,
		}
	}
	return result
}
