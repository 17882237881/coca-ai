package llm

import "context"

// Message 表示一条对话消息
type Message struct {
	Role    string // "user", "assistant", "system"
	Content string
}

// StreamCallback 流式响应回调函数
// delta: 本次返回的文本片段
// 返回 error 可以中断流式传输
type StreamCallback func(delta string) error

// ChatClient LLM 客户端接口
// 抽象接口便于未来切换模型 (如从通义千问切换到 GPT-4 或 DeepSeek)
type ChatClient interface {
	// Chat 普通对话 (非流式)
	// 返回完整的 AI 回复内容
	Chat(ctx context.Context, messages []Message) (string, error)

	// StreamChat 流式对话
	// 通过 callback 逐步返回生成的内容
	StreamChat(ctx context.Context, messages []Message, callback StreamCallback) error

	// Summarize 生成摘要
	// 将历史消息压缩成简短摘要
	Summarize(ctx context.Context, messages []Message) (string, error)
}
