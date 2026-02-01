# Coca AI (v1.0) 🚀

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/17882237881/coca-ai)](https://goreportcard.com/report/github.com/17882237881/coca-ai)
[![Vue 3](https://img.shields.io/badge/vue-3.x-green.svg)](https://vuejs.org/)

**Coca AI** 是一个高性能、现代化的 AI 对话系统，基于 **Go (Backend)** 和 **Vue 3 (Frontend)** 构建。它不仅拥有类似 ChatGPT 的流畅 UI 和流式响应体验，更在后端采用了企业级的高并发架构设计。

![Architecture](docs/architecture_diagram.md)

## ✨ 核心特性 (v1.0)

### 🤖 智能对话
- **ChatGPT 风格 UI**: 深色主题，打字机流式效果 (SSE)，支持 Markdown 渲染和代码高亮。
- **多会话管理**: 支持创建多个会话，历史记录自动保存。
- **大模型集成**: 接入通义千问 (Qwen) API，支持智能上下文理解。

### ⚡ 高性能架构
- **异步削峰**: 采用 **Write-Behind** 模式。消息先写 Redis 缓存 + 投递 Kafka，再异步落库 MySQL，极大降低接口延迟。
- **分层设计**: 严格遵循 DDD (领域驱动设计) 分层架构。
- **依赖注入**: 使用 Google Wire 进行依赖注入，代码解耦。

### 📊 可观测性 (Observability)
- **Prometheus**: 内置 Metrics 监控 (QPS, Latency, Goroutines)。
- **Jaeger**: 全链路分布式追踪 (HTTP -> Service -> Redis/Kafka -> MySQL)。

## 🛠 技术栈

| 领域 | 技术选型 | 说明 |
|------|----------|------|
| **Frontend** | Vue 3, TypeScript, Vite | TailwindCSS 样式, marked 解析 |
| **Backend** | Go 1.22+, Gin | GORM, Wire, Viper, Zap |
| **Messaging** | Kafka, Zookeeper | segmentio/kafka-go 客户端 |
| **Cache** | Redis 7.0 | 会话缓存, Write-Through |
| **Database** | MySQL 8.0 | 消息持久化 |
| **DevOps** | Docker, Nginx | 多阶段构建, Prometheus, Jaeger |

## 🚀 快速开始

### 前置要求
- Docker & Docker Compose
- Go 1.22+ (可选，本地开发需要)
- Node.js 18+ (可选，本地开发需要)

### 本地一键启动 (推荐)

1. **克隆仓库**
   ```bash
   git clone https://github.com/17882237881/coca-ai.git
   cd coca-ai
   ```

2. **配置环境变量**
   复制 `configs/config.yaml.example` 为 `configs/config.yaml`，并填入你的 LLM API Key。

3. **启动服务**
   ```bash
   # 启动所有依赖 (MySQL, Redis, Kafka, Nginx, Backend)
   docker-compose -f deploy/docker-compose.prod.yml up -d
   ```

4. **访问应用**
   - **Frontend**: http://localhost
   - **Jaeger UI**: http://localhost:16686
   - **Prometheus**: http://localhost:9090

### 开发模式启动

**后端**:
```bash
# 启动基础依赖 (MySQL, Redis, Kafka)
docker-compose up -d mysql redis kafka zookeeper

# 运行后端
go run cmd/server/main.go
```

**前端**:
```bash
cd web
npm install
npm run dev
```

## 📂 项目结构

```
coca-ai/
├── cmd/                # 应用程序入口
├── configs/            # 配置文件
├── deploy/             # Docker Compose 部署文件
├── docs/               # 文档 (Kafka 教程, 架构图)
├── internal/
│   ├── api/            # DTO 对象
│   ├── config/         # 配置加载
│   ├── domain/         # 领域实体
│   ├── handler/        # HTTP 路由处理
│   ├── ioc/            # 依赖注入 (Wire)
│   ├── mq/             # 消息队列 (Kafka)
│   ├── repository/     # 数据访问层 (DAO)
│   └── service/        # 业务逻辑层
├── web/                # Vue 3 前端项目
└── go.mod
```

## 📚 文档资源

- [Kafka 核心概念与架构设计](docs/kafka_learning/01_kafka_concepts.md)
- [Coca AI 异步消息架构详解](docs/kafka_learning/02_coca_ai_usage.md)
- [代码实现细节](docs/kafka_learning/03_code_implementation.md)
- [运维命令手册](docs/kafka_learning/04_operations_and_commands.md)

## 📅 Roadmap

- [x] v1.0: 基础对话, 异步消息, 监控体系。
- [ ] v1.x: Python 向量服务集成 (RAG)。
- [ ] v2.0: 知识库管理后台, 多模型支持。

---
Built with ❤️ by Coca AI Team.
