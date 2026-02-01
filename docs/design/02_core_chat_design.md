# 基础对话通道 - 系统设计文档

**模块**: 第三阶段 - AI 对话核心  
**功能**: 基础对话通道  
**版本**: v1.0  
**日期**: 2026-01-30

---

## 1. 系统架构概览

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              Frontend (Vue3)                                │
│                         /chat/sessions/:id/messages                         │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │ SSE
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                              API Gateway (Gin)                              │
│                              ChatHandler                                    │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
          ┌─────────────────────────┼─────────────────────────┐
          ▼                         ▼                         ▼
┌──────────────────┐    ┌──────────────────┐    ┌──────────────────┐
│   ChatService    │    │  ContextService  │    │   KafkaProducer  │
│   (核心对话逻辑)  │    │  (上下文构建)     │    │   (异步发送)      │
└──────────────────┘    └──────────────────┘    └──────────────────┘
          │                         │
          ▼                         ▼
┌──────────────────┐    ┌──────────────────┐
│   Eino ChatModel │    │     Redis        │
│   (通义千问)      │    │   (消息缓存)      │
└──────────────────┘    └──────────────────┘
                                    
┌─────────────────────────────────────────────────────────────────────────────┐
│                           Kafka Consumer (独立协程)                          │
│                           消费 chat.messages Topic                          │
│                                    │                                        │
│                                    ▼                                        │
│                              MySQL 持久化                                    │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 2. 目录结构设计

```
internal/
├── domain/
│   ├── session.go          # Session 领域模型
│   └── message.go          # Message 领域模型
├── handler/
│   └── chat.go             # ChatHandler (HTTP 接口)
├── service/
│   ├── chat.go             # ChatService (核心业务逻辑)
│   └── context.go          # ContextService (上下文构建)
├── repository/
│   ├── session.go          # SessionRepository
│   ├── message.go          # MessageRepository
│   └── dao/
│       ├── session.go      # Session DAO (MySQL)
│       └── message.go      # Message DAO (MySQL)
├── llm/
│   ├── client.go           # LLM 客户端接口
│   └── qwen.go             # 通义千问实现 (via Eino)
├── mq/
│   ├── producer.go         # Kafka Producer
│   └── consumer.go         # Kafka Consumer
└── ioc/
    ├── llm.go              # LLM 初始化
    └── kafka.go            # Kafka 初始化
```

---

## 3. 数据库设计

### 3.1 sessions 表

```sql
CREATE TABLE sessions (
    id          BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    user_id     BIGINT UNSIGNED NOT NULL,
    title       VARCHAR(255) NOT NULL DEFAULT '',
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_user_id (user_id),
    INDEX idx_updated_at (updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 3.2 messages 表

```sql
CREATE TABLE messages (
    id          BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    session_id  BIGINT UNSIGNED NOT NULL,
    role        VARCHAR(20) NOT NULL,  -- 'user', 'assistant', 'system'
    content     TEXT NOT NULL,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_session_id (session_id),
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

---

## 4. API 设计

### 4.1 创建会话

```
POST /chat/sessions
Authorization: Bearer <access_token>

Response 200:
{
    "code": 200,
    "data": {
        "session_id": 12345,
        "title": "New Chat"
    }
}
```

### 4.2 获取会话列表

```
GET /chat/sessions
Authorization: Bearer <access_token>

Response 200:
{
    "code": 200,
    "data": [
        {"session_id": 12345, "title": "关于Go的问题", "updated_at": "2026-01-30T12:00:00Z"},
        {"session_id": 12344, "title": "Docker部署", "updated_at": "2026-01-29T18:00:00Z"}
    ]
}
```

### 4.3 获取历史消息

```
GET /chat/sessions/:id/messages
Authorization: Bearer <access_token>

Response 200:
{
    "code": 200,
    "data": [
        {"role": "user", "content": "你好", "created_at": "..."},
        {"role": "assistant", "content": "你好！有什么可以帮您的吗？", "created_at": "..."}
    ]
}
```

### 4.4 发送消息 (SSE 流式响应)

```
POST /chat/sessions/:id/messages
Authorization: Bearer <access_token>
Content-Type: application/json
Accept: text/event-stream

Request:
{
    "content": "请解释一下什么是 DDD"
}

Response (SSE Stream):
event: message
data: {"delta": "DDD"}

event: message
data: {"delta": "(领域驱动设计)"}

event: message
data: {"delta": "是一种软件设计方法..."}

event: done
data: {"message_id": 67890}
```

### 4.5 删除会话

```
DELETE /chat/sessions/:id
Authorization: Bearer <access_token>

Response 200:
{
    "code": 200,
    "msg": "Session deleted"
}
```

---

## 5. 核心流程设计

### 5.1 发送消息时序图

```
User          Handler        ChatService      ContextService     LLM(Eino)      Redis         Kafka
  │              │                │                 │               │              │             │
  │─POST message─▶│                │                 │               │              │             │
  │              │──存储用户消息──▶│                 │               │              │             │
  │              │                │───写入Redis────▶│               │              │             │
  │              │                │◀───────────────│               │──────────────▶│             │
  │              │                │───发送Kafka───▶│               │              │             │──────▶│
  │              │                │                 │               │              │             │
  │              │                │──构建上下文────▶│               │              │             │
  │              │                │◀──返回Prompt───│               │              │             │
  │              │                │                 │               │              │             │
  │              │                │──Stream调用───▶│               │              │             │
  │              │                │◀──流式Token────│               │              │             │
  │◀──SSE推送────│◀───────────────│                 │               │              │             │
  │              │                │                 │               │              │             │
  │              │                │──存储AI回复───▶│               │              │             │
  │              │                │               (同上写Redis+Kafka)              │             │
```

### 5.2 上下文构建流程

```go
// ContextService.BuildContext 伪代码
func (s *ContextService) BuildContext(sessionID int64, userMsg string) []Message {
    // 1. 获取最近 K 条消息 (从 Redis)
    recentMessages := s.redis.GetRecentMessages(sessionID, K=10)
    
    // 2. 获取历史消息数量
    totalCount := s.redis.GetMessageCount(sessionID)
    
    // 3. 如果历史超过阈值，生成摘要
    var summary string
    if totalCount > K {
        olderMessages := s.repo.GetOlderMessages(sessionID, beforeID)
        summary = s.llm.Summarize(olderMessages)  // 调用 LLM 生成摘要
    }
    
    // 4. 组装 Context
    context := []Message{
        {Role: "system", Content: SYSTEM_PROMPT},
    }
    if summary != "" {
        context = append(context, Message{Role: "system", Content: "[历史摘要] " + summary})
    }
    context = append(context, recentMessages...)
    context = append(context, Message{Role: "user", Content: userMsg})
    
    return context
}
```

---

## 6. Eino 集成设计

### 6.1 LLM 客户端接口

```go
// internal/llm/client.go
package llm

import "context"

type Message struct {
    Role    string
    Content string
}

type StreamCallback func(delta string) error

type ChatClient interface {
    // Chat 普通对话 (非流式)
    Chat(ctx context.Context, messages []Message) (string, error)
    
    // StreamChat 流式对话
    StreamChat(ctx context.Context, messages []Message, callback StreamCallback) error
    
    // Summarize 生成摘要
    Summarize(ctx context.Context, messages []Message) (string, error)
}
```

### 6.2 通义千问实现

```go
// internal/llm/qwen.go
package llm

import (
    "context"
    "github.com/cloudwego/eino-ext/components/model/openai"
    "github.com/cloudwego/eino/components/model"
)

type QwenClient struct {
    chatModel model.ChatModel
}

func NewQwenClient(apiKey, baseURL string) (*QwenClient, error) {
    // 使用 Eino 的 OpenAI 兼容接口连接通义千问
    chatModel, err := openai.NewChatModel(context.Background(), &openai.ChatModelConfig{
        BaseURL: baseURL,  // 通义千问 API 端点
        APIKey:  apiKey,
        Model:   "qwen-plus",
    })
    if err != nil {
        return nil, err
    }
    return &QwenClient{chatModel: chatModel}, nil
}

func (c *QwenClient) StreamChat(ctx context.Context, messages []Message, callback StreamCallback) error {
    // 转换消息格式
    einoMessages := convertToEinoMessages(messages)
    
    // 调用流式接口
    stream, err := c.chatModel.Stream(ctx, einoMessages)
    if err != nil {
        return err
    }
    
    // 读取流式响应
    for chunk := range stream {
        if err := callback(chunk.Content); err != nil {
            return err
        }
    }
    return nil
}
```

---

## 7. Kafka 设计

### 7.1 Topic 定义

| Topic | 描述 | 分区数 | 保留策略 |
|-------|------|--------|---------|
| `chat.messages` | 聊天消息 | 3 | 7天 |

### 7.2 消息格式

```json
{
    "session_id": 12345,
    "message_id": 67890,
    "role": "user",
    "content": "你好",
    "user_id": 1001,
    "created_at": "2026-01-30T12:00:00Z"
}
```

### 7.3 Producer

```go
// internal/mq/producer.go
type MessageProducer interface {
    SendMessage(ctx context.Context, msg *MessageEvent) error
}

type KafkaProducer struct {
    writer *kafka.Writer
}

func (p *KafkaProducer) SendMessage(ctx context.Context, msg *MessageEvent) error {
    data, _ := json.Marshal(msg)
    return p.writer.WriteMessages(ctx, kafka.Message{
        Key:   []byte(strconv.FormatInt(msg.SessionID, 10)),
        Value: data,
    })
}
```

### 7.4 Consumer

```go
// internal/mq/consumer.go
type MessageConsumer struct {
    reader  *kafka.Reader
    msgRepo repository.MessageRepository
}

func (c *MessageConsumer) Start(ctx context.Context) {
    for {
        msg, err := c.reader.ReadMessage(ctx)
        if err != nil {
            continue
        }
        
        var event MessageEvent
        json.Unmarshal(msg.Value, &event)
        
        // 落库 MySQL
        c.msgRepo.Save(ctx, &domain.Message{
            SessionID: event.SessionID,
            Role:      event.Role,
            Content:   event.Content,
            CreatedAt: event.CreatedAt,
        })
    }
}
```

---

## 8. Redis 缓存设计

### 8.1 Key 设计

| Key Pattern | 类型 | TTL | 描述 |
|-------------|------|-----|------|
| `chat:session:{session_id}:messages` | List | 24h | 会话消息列表 |
| `chat:session:{session_id}:count` | String | 24h | 消息数量 |

### 8.2 操作

```go
// 添加消息
RPUSH chat:session:12345:messages "{...json...}"
INCR  chat:session:12345:count

// 获取最近 K 条
LRANGE chat:session:12345:messages -10 -1
```

---

## 9. 依赖配置

### 9.1 环境变量

| 变量名 | 描述 | 示例值 |
|--------|------|--------|
| `QWEN_API_KEY` | 通义千问 API Key | `sk-xxx` |
| `QWEN_BASE_URL` | 通义千问 API 端点 | `https://dashscope.aliyuncs.com/compatible-mode/v1` |
| `KAFKA_BROKERS` | Kafka Broker 地址 | `localhost:9092` |

### 9.2 Go 依赖

```bash
go get github.com/cloudwego/eino
go get github.com/cloudwego/eino-ext/components/model/openai
go get github.com/segmentio/kafka-go
```

---

## 10. 实现优先级

| 阶段 | 内容 | 依赖 |
|------|------|------|
| Phase 1 | Session/Message 领域模型 + DAO | - |
| Phase 2 | Redis 缓存层 | Phase 1 |
| Phase 3 | Eino + 通义千问集成 | - |
| Phase 4 | SSE 流式接口 | Phase 1, 2, 3 |
| Phase 5 | Kafka Producer/Consumer | Phase 1 |
| Phase 6 | 智能上下文 (摘要压缩) | Phase 3, 4 |

---

## 11. 详细实现步骤 (Step-by-Step Implementation)

**核心原则**: 先定契约 (API)，后定实现 (Domain/Dao)。与前端对齐接口后再写业务逻辑。

### 步骤 1: 定义 API 契约 (Interface Definition)
*   **动作**: 编写 `internal/handler/chat.go` 的结构体定义部分。
*   **Why**: **API First**。先和前端确认"你要传什么给我，我返回什么给你"，避免后端写完了接口不对的连锁反应。
*   **代码内容**:
    ```go
    // internal/handler/chat.go
    type ChatHandler struct {
        chatSvc service.ChatService
    }

    type CreateSessionResp struct {
        SessionID int64  `json:"session_id"`
        Title     string `json:"title"`
    }

    type SendMessageReq struct {
        Content string `json:"content" binding:"required"`
    }

    type SessionListItem struct {
        SessionID int64  `json:"session_id"`
        Title     string `json:"title"`
        UpdatedAt string `json:"updated_at"`
    }
    ```

### 步骤 2: 定义领域实体 (Domain Entity)
*   **动作**: 在 `internal/domain` 创建 `session.go` 和 `message.go`。
*   **Why**: 领域层是连接"外部接口"和"底层存储"的桥梁，代表核心业务对象。
*   **代码内容**:
    ```go
    // internal/domain/session.go
    type Session struct {
        ID        int64
        UserID    int64
        Title     string
        CreatedAt time.Time
        UpdatedAt time.Time
    }

    // internal/domain/message.go
    type Message struct {
        ID        int64
        SessionID int64
        Role      string  // "user", "assistant", "system"
        Content   string
        CreatedAt time.Time
    }
    ```

### 步骤 3: 数据库设计与 DAO 实现
*   **动作**: 在 `internal/repository/dao` 创建 `session.go` 和 `message.go`。
*   **Why**: 根据 MySQL 的特性设计表结构，DAO 层负责把 Domain 对象变成数据库能存的 Row。
*   **要点**:
    - GORM 自动迁移
    - 索引优化 (user_id, session_id)
    - 外键约束 (session -> messages 级联删除)

### 步骤 4: 实现 Repository 层
*   **动作**: 在 `internal/repository` 创建 `session.go` 和 `message.go`。
*   **Why**: 隔离层。Service 层只认 Domain，不应该知道底下用的是 MySQL 还是 Redis。
*   **要点**:
    - `SessionRepository`: Create, FindByUserID, Delete
    - `MessageRepository`: Create, FindBySessionID, Save (供 Consumer 使用)

### 步骤 5: 初始化 LLM 客户端 (Eino + 通义千问)
*   **动作**: 在 `internal/llm` 创建 `client.go` (接口) 和 `qwen.go` (实现)。
*   **Why**: 抽象接口便于未来切换模型 (如切换到 GPT-4 或 DeepSeek)。
*   **依赖**: `github.com/cloudwego/eino-ext/components/model/openai`
*   **环境变量**: `QWEN_API_KEY`, `QWEN_BASE_URL`

### 步骤 6: 实现 Redis 缓存层
*   **动作**: 在 `internal/repository/message.go` 中添加 Redis 读写逻辑。
*   **Why**: 热数据缓存，提升上下文构建的读取性能。
*   **要点**:
    - 写入时: `RPUSH` 添加消息到 List
    - 读取时: `LRANGE` 获取最近 K 条
    - TTL: 24 小时

### 步骤 7: 实现 Kafka Producer
*   **动作**: 在 `internal/mq` 创建 `producer.go`。
*   **Why**: 异步解耦写入操作，避免 MySQL 成为性能瓶颈。
*   **依赖**: `github.com/segmentio/kafka-go`
*   **Topic**: `chat.messages`

### 步骤 8: 实现 Kafka Consumer
*   **动作**: 在 `internal/mq` 创建 `consumer.go`。
*   **Why**: 消费 Kafka 消息，异步落库到 MySQL。
*   **启动方式**: 服务启动时，开启独立 goroutine 运行 Consumer。

### 步骤 9: 实现 ChatService 核心业务逻辑
*   **动作**: 在 `internal/service` 创建 `chat.go`。
*   **核心方法**:
    - `CreateSession(ctx, userID)`: 创建新会话
    - `SendMessage(ctx, sessionID, content, stream)`: 发送消息并流式回复
    - `GetSessionList(ctx, userID)`: 获取会话列表
    - `GetMessages(ctx, sessionID)`: 获取会话历史
*   **流程**:
    1. 将用户消息写入 Redis
    2. 发送 Kafka 消息 (异步落库)
    3. 调用 ContextService 构建上下文
    4. 调用 LLM StreamChat
    5. 将 AI 回复写入 Redis + Kafka

### 步骤 10: 实现 ContextService 上下文构建
*   **动作**: 在 `internal/service` 创建 `context.go`。
*   **Why**: 单独抽取上下文构建逻辑，便于后续升级 (如加入 RAG)。
*   **核心逻辑**:
    1. 获取最近 K 条消息
    2. 如果历史超过阈值，调用 LLM 生成摘要
    3. 组装 System Prompt + 摘要 + 最近消息 + 用户输入

### 步骤 11: 实现 SSE 流式响应
*   **动作**: 在 `internal/handler/chat.go` 实现 `SendMessage` 方法。
*   **Why**: SSE 是浏览器原生支持的流式协议，实现打字机效果。
*   **Gin SSE 实现**:
    ```go
    func (h *ChatHandler) SendMessage(c *gin.Context) {
        c.Header("Content-Type", "text/event-stream")
        c.Header("Cache-Control", "no-cache")
        
        h.chatSvc.SendMessage(ctx, sessionID, req.Content, func(delta string) error {
            c.SSEvent("message", gin.H{"delta": delta})
            c.Writer.Flush()
            return nil
        })
        
        c.SSEvent("done", gin.H{"message_id": msgID})
    }
    ```

### 步骤 12: 依赖注入 (Wire)
*   **动作**: 更新 `cmd/server/wire.go`，将所有新组件串起来。
*   **新增 Provider**:
    - `ioc.InitLLMClient`
    - `ioc.InitKafkaProducer`
    - `ioc.InitKafkaConsumer`
    - `handler.NewChatHandler`
    - `service.NewChatService`
    - `service.NewContextService`
    - `repository.NewSessionRepository`
    - `repository.NewMessageRepository`

---

## 12. 验证计划

### 12.1 手动测试

```bash
# 1. 创建会话
curl -X POST http://localhost:8080/chat/sessions \
  -H "Authorization: Bearer <token>"

# 2. 发送消息 (SSE)
curl -N -X POST http://localhost:8080/chat/sessions/1/messages \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"content": "你好"}'

# 3. 获取历史
curl http://localhost:8080/chat/sessions/1/messages \
  -H "Authorization: Bearer <token>"
```

### 12.2 验收 Checklist

- [ ] 创建会话返回 session_id
- [ ] 发送消息后，SSE 流式推送 AI 回复
- [ ] Redis 中有会话消息缓存
- [ ] MySQL 中异步落库成功 (Kafka Consumer)
- [ ] 历史超过 10 条时，摘要正确生成
