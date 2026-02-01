# Eino 框架使用指南

**版本**: v1.0  
**更新日期**: 2026-01-31  
**说明**: 本文档介绍 Eino 框架在 Coca AI 项目中的使用方法，随着项目开发会持续更新。

---

## 1. Eino 简介

**Eino** 是字节跳动 (ByteDance) 开源的 **Go 语言 LLM 应用开发框架**，类似于 Python 的 LangChain，但更符合 Go 的编程范式。

**核心特点**:
- 组件化架构，便于扩展和替换
- 强类型检查，编译时发现错误
- 原生支持流式响应
- 已在豆包、抖音等产品中验证

**GitHub**: https://github.com/cloudwego/eino

---

## 2. 安装依赖

```bash
# 核心库
go get github.com/cloudwego/eino

# OpenAI 兼容组件 (用于通义千问)
go get github.com/cloudwego/eino-ext/components/model/openai
```

---

## 3. 核心概念

### 3.1 ChatModel

`ChatModel` 是与 LLM 交互的核心组件，定义了两个主要方法：

| 方法 | 描述 |
|------|------|
| `Generate(ctx, messages)` | 普通对话，返回完整响应 |
| `Stream(ctx, messages)` | 流式对话，逐 Token 返回 |

### 3.2 schema.Message

Eino 使用 `schema.Message` 表示对话消息：

```go
import "github.com/cloudwego/eino/schema"

message := &schema.Message{
    Role:    schema.User,      // 角色类型
    Content: "你好",           // 消息内容
}
```

**角色类型 (schema.RoleType)**:
| 常量 | 描述 |
|------|------|
| `schema.User` | 用户消息 |
| `schema.Assistant` | AI 助手回复 |
| `schema.System` | 系统提示 (System Prompt) |

---

## 4. 项目中的使用示例

### 4.1 创建 ChatModel 客户端

```go
import (
    "context"
    "github.com/cloudwego/eino-ext/components/model/openai"
)

// 使用 OpenAI 兼容接口连接通义千问
chatModel, err := openai.NewChatModel(context.Background(), &openai.ChatModelConfig{
    BaseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1",  // 通义千问 API 端点
    APIKey:  "sk-xxx",                                             // API Key
    Model:   "qwen-plus",                                          // 模型名称
})
```

**关键参数说明**:
- `BaseURL`: 通义千问提供 OpenAI 兼容的 API 端点，因此可以用 Eino 的 OpenAI 组件
- `Model`: 支持 `qwen-plus`, `qwen-turbo`, `qwen-max` 等

### 4.2 普通对话 (Generate)

```go
// 构建消息列表
messages := []*schema.Message{
    {Role: schema.System, Content: "你是一个友好的助手。"},
    {Role: schema.User, Content: "你好，介绍一下自己。"},
}

// 调用 Generate 获取完整响应
response, err := chatModel.Generate(ctx, messages)
if err != nil {
    return err
}

fmt.Println(response.Content)  // 输出 AI 回复
```

**返回值**:
- `response` 是 `*schema.Message` 类型
- `response.Content` 是 AI 回复的文本内容

### 4.3 流式对话 (Stream)

```go
import "io"

// 调用 Stream 获取流式响应
stream, err := chatModel.Stream(ctx, messages)
if err != nil {
    return err
}
defer stream.Close()  // 重要：记得关闭流

// 循环读取每个 chunk
for {
    chunk, err := stream.Recv()
    if err == io.EOF {
        break  // 流结束
    }
    if err != nil {
        return err
    }
    
    // chunk.Content 是本次返回的文本片段
    fmt.Print(chunk.Content)  // 逐字打印，实现打字机效果
}
```

**关键点**:
- `stream.Recv()` 每次返回一个 chunk（文本片段）
- 当返回 `io.EOF` 表示流结束
- **必须调用 `stream.Close()`** 释放资源

---

## 5. 完整代码示例

以下是项目中 `internal/llm/qwen.go` 的关键代码：

```go
package llm

import (
    "context"
    "io"
    
    "github.com/cloudwego/eino-ext/components/model/openai"
    "github.com/cloudwego/eino/schema"
)

type QwenClient struct {
    chatModel *openai.ChatModel
}

// 创建客户端
func NewQwenClient(apiKey, baseURL string) (*QwenClient, error) {
    chatModel, err := openai.NewChatModel(context.Background(), &openai.ChatModelConfig{
        BaseURL: baseURL,
        APIKey:  apiKey,
        Model:   "qwen-plus",
    })
    if err != nil {
        return nil, err
    }
    return &QwenClient{chatModel: chatModel}, nil
}

// 流式对话
func (c *QwenClient) StreamChat(ctx context.Context, messages []Message, callback func(string) error) error {
    // 1. 转换消息格式
    einoMessages := make([]*schema.Message, len(messages))
    for i, msg := range messages {
        einoMessages[i] = &schema.Message{
            Role:    toEinoRole(msg.Role),
            Content: msg.Content,
        }
    }
    
    // 2. 调用流式接口
    stream, err := c.chatModel.Stream(ctx, einoMessages)
    if err != nil {
        return err
    }
    defer stream.Close()
    
    // 3. 读取流式响应
    for {
        chunk, err := stream.Recv()
        if err == io.EOF {
            break
        }
        if err != nil {
            return err
        }
        
        // 4. 通过回调函数传递每个 chunk
        if err := callback(chunk.Content); err != nil {
            return err
        }
    }
    return nil
}

// 辅助函数：转换角色类型
func toEinoRole(role string) schema.RoleType {
    switch role {
    case "user":
        return schema.User
    case "assistant":
        return schema.Assistant
    case "system":
        return schema.System
    default:
        return schema.User
    }
}
```

---

## 6. 常见问题

### Q1: 为什么用 OpenAI 组件连接通义千问？
通义千问提供了 OpenAI 兼容的 API 端点，格式与 OpenAI 一致，因此可以直接使用 Eino 的 OpenAI 组件。

### Q2: 如何切换到其他模型 (如 GPT-4)？
只需修改 `BaseURL` 和 `APIKey`：
```go
chatModel, _ := openai.NewChatModel(ctx, &openai.ChatModelConfig{
    BaseURL: "https://api.openai.com/v1",  // OpenAI 官方端点
    APIKey:  "sk-openai-xxx",
    Model:   "gpt-4",
})
```

### Q3: Stream 返回的 chunk 是什么？
每个 chunk 是一个 `*schema.Message`，其 `Content` 字段包含本次返回的文本片段（通常是几个字或一个词）。

---

## 7. 待补充内容

以下功能将在后续开发中使用并补充文档：

- [ ] Tool Calling (工具调用)
- [ ] Retriever (RAG 检索)
- [ ] ChatTemplate (Prompt 模板)
- [ ] Agent 编排
- [ ] 可视化调试工具
