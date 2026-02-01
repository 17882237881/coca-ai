# Coca AI 技术架构图 (Current State)

本文档展示了 Coca AI 项目截止目前的总体技术架构。

## 系统架构总览

```mermaid
graph TD
    %% User Layer
    User((User))
    
    %% Frontend Layer
    subgraph "Frontend Layer (Web)"
        Nginx[Nginx Reverse Proxy]
        Vue[Vue 3 SPA]
        Vite[Vite Build Tool]
        SSE_Client[SSE Client]
    end

    %% Backend Layer
    subgraph "Backend Layer (Go/Gin)"
        API_Gateway[Gin HTTP Server]
        Middleware[Middleware: CORS, JWT, Recovery]
        
        subgraph "Domain Services"
            UserService[User Service]
            ChatService[Chat Service]
            ContextService[Context Builder]
        end
        
        subgraph "Async Processing"
            AsyncProducer[Kafka Producer]
            AsyncConsumer[Kafka Consumer]
        end
        
        subgraph "Observability SDKs"
            PromSDK[Prometheus SDK]
            OtelSDK[OpenTelemetry (Jaeger)]
        end
    end

    %% Data & Infrastructure Layer
    subgraph "Infrastructure & Data"
        MySQL[(MySQL 8.0)]
        Redis[(Redis 7.0)]
        Kafka{Kafka Message Queue}
        Zookeeper[Zookeeper]
        Prometheus[Prometheus Server]
        Jaeger[Jaeger Collector]
    end

    %% External Services
    LLM_API["☁️ Tongyi Qianwen LLM API"]

    %% Connections
    User -->|HTTP/WebSocket| Nginx
    Nginx -->|Static Files| Vue
    Nginx -->|Proxy /api| API_Gateway
    
    Vue -->|REST API| API_Gateway
    Vue -->|EventSource (SSE)| API_Gateway

    API_Gateway --> Middleware
    Middleware --> UserService
    Middleware --> ChatService

    UserService -->|Read/Write| MySQL
    UserService -->|Cache| Redis

    ChatService -->|1. Get Context| ContextService
    ChatService -->|2. Stream Chat| LLM_API
    ChatService -->|3. Cache Msg| Redis
    ChatService -->|4. Async Save| AsyncProducer

    AsyncProducer -->|Publish| Kafka
    Kafka -->|Store| Zookeeper
    Kafka -->|Consume| AsyncConsumer
    AsyncConsumer -->|Persist| MySQL

    %% Observability Flows
    API_Gateway -.->|Metrics| PromSDK
    PromSDK -.->|Scrape| Prometheus
    
    API_Gateway -.->|Trace| OtelSDK
    ChatService -.->|Trace| OtelSDK
    AsyncConsumer -.->|Trace| OtelSDK
    OtelSDK -.->|Export| Jaeger
```

## 核心数据流说明

1.  **用户认证流**: 
    -   `Login API` -> `UserService` -> 校验 Redis (Session) -> 校验 MySQL -> 签发 JWT -> 返回 Token。

2.  **实时对话流 (SSE)**:
    -   前端 `EventSource` 连接 `/chat/stream`。
    -   后端 `ChatService` 构建上下文 -> 调用 LLM 流式接口。
    -   LLM 每生成一个 Token -> 后端通过 SSE 实时推送到前端。

3.  **异步持久化流 (Write-Behind)**:
    -   AI 生成完整的回答后 -> `ChatService` 将完整消息写入 Redis 缓存 (快速响应) -> 同时投递到 Kafka `chat.messages` Topic。
    -   后台 Consumer 异步消费 -> 写入 MySQL (削峰填谷，不阻塞主线程)。

4.  **可观测性流**:
    -   **Metrics**: Prometheus 每 15s 拉取一次 `/metrics`，监控 QPS、Goroutines、内存。
    -   **Tracing**: 每个请求生成 TraceID，Jaeger 记录从 HTTP 入口到 Database/LLM 的全链路耗时。
