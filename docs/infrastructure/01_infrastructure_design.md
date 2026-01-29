# 基础设施设计文档 - Coca AI

**版本**: 1.0
**状态**: 草稿
**作者**: Antigravity & User

## 1. 概述
本文档概述了 **Coca AI** 本地开发环境的基础设施需求与设计。目标是使用 Docker Compose 建立一个支持高并发、高可用的基础环境。

## 2. 基础设施拓扑

系统包含由 `docker-compose` 编排的以下核心服务。

| 服务类别 | 服务名称 | 镜像 / 版本 | 端口 (宿主机:容器) | 用途 | 数据卷 |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **数据库** | `coca-mysql` | `mysql:8.0` | `13306:3306` | 主要关系型数据存储 (用户, 元数据)。 | `./infra/data/mysql` |
| **缓存** | `coca-redis` | `redis:7.0` | `16379:6379` | 会话存储, 限流, 缓存。 | `./infra/data/redis` |
| **向量库** | `coca-milvus` | `milvusdb/milvus:v2.3.x` | `19530:19530` <br> `9091:9091` | 用于 RAG 的向量嵌入存储。 | `./infra/data/milvus` |
| **消息队列** | `coca-kafka` | `confluentinc/cp-kafka:7.4.0` | `9092:9092` | 异步日志, 聊天历史缓冲。 | `./infra/data/kafka` |
| **MQ 协调** | `coca-zookeeper`| `confluentinc/cp-zookeeper:7.4.0` | `2181:2181` | Kafka 协调服务。 | `./infra/data/zookeeper` |
| **链路追踪** | `coca-jaeger` | `jaegertracing/all-in-one`| `16686:16686` (UI) <br> `4317:4317` (OTLP) | 分布式链路追踪 (OpenTelemetry)。 | N/A |
| **监控** | `coca-prometheus`| `prom/prometheus` | `19090:9090` | 系统指标采集。 | `./infra/data/prometheus` |

## 3. 详细配置

### 3.1 网络
- **网络名称**: `coca_network`
- **驱动**: `bridge`
- 所有服务都处于该网络中，允许内部 DNS 解析 (例如后端应用可直接访问 `coca-mysql`)。

### 3.2 存储策略
- 所有持久化数据映射到宿主机目录 `./infra/data/`，确保容器重启后数据不丢失。
- **Git 忽略**: `infra/data/` 目录必须加入 `.gitignore`，防止提交数据库文件。

### 3.3 服务详情

#### MySQL
- **环境变量**:
    - `MYSQL_ROOT_PASSWORD`: `root` (仅限本地开发)
    - `MYSQL_DATABASE`: `coca_db`
- **配置**: 默认字符集 `utf8mb4`。

#### Milvus (单机版)
- 需要 `etcd` 和 `minio` 作为依赖。
- **依赖**:
    - `coca-etcd`: Milvus 元数据存储。
    - `coca-minio`: 对象存储 (兼容 S3)，用于存储 Milvus 日志/文件。
- **复杂度**: Milvus 单机版涉及 3 个容器 (Milvus, Etcd, Minio)。

#### Kafka
- **环境变量**:
    - `KAFKA_CFG_ZOOKEEPER_CONNECT`: `coca-zookeeper:2181`
    - `KAFKA_CFG_ADVERTISED_LISTENERS`: `PLAINTEXT://localhost:9092` (外部), `CLIENT://coca-kafka:9093` (内部)
    - `ALLOW_PLAINTEXT_LISTENER`: `yes`

### 4. 验证计划
- **端口检查**: 确保所有暴露端口都在监听。
- **连接性检查**: 使用数据库工具 (DBeaver, RedisManager) 连接 localhost 端口。
- **Milvus 健康度**: 检查 Milvus 健康端点或使用 Attu (Milvus GUI，可选但推荐)。
