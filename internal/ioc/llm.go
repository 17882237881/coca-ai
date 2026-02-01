package ioc

import (
	"coca-ai/internal/config"
	"coca-ai/internal/llm"
)

// InitLLMClient 初始化 LLM 客户端
func InitLLMClient() llm.ChatClient {
	cfg := config.Get()

	if cfg.LLM.APIKey == "" {
		panic("LLM API Key is not configured. Set llm.api_key in config.yaml or QWEN_API_KEY environment variable.")
	}

	client, err := llm.NewQwenClient(&llm.QwenConfig{
		APIKey:  cfg.LLM.APIKey,
		BaseURL: cfg.LLM.BaseURL,
		Model:   cfg.LLM.Model,
	})
	if err != nil {
		panic("Failed to create LLM client: " + err.Error())
	}

	return client
}
