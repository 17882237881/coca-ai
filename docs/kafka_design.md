# Kafka 设计说明（Coca-AI）

## 1. 角色与目标
Kafka 在本项目中是**异步消息队列**，核心目标：
- **削峰填谷**：高并发下将写入 MySQL 的压力转移到 Kafka。
- **解耦与容错**：业务请求与持久化解耦，数据库抖动不影响前端响应。
- **消息可靠性**：保证消息不丢、可重试、可追踪。

相关实现位置：
- Producer：`internal/mq/producer.go`
- Consumer：`internal/mq/consumer.go`
- Handler：`internal/mq/handler.go`
- 注入与启动：`internal/ioc/kafka.go`, `cmd/server/app.go`, `cmd/server/wire_gen.go`
- 配置：`internal/config/config.go`, `configs/config.yaml`
- 部署：`deploy/docker-compose.prod.yml`

## 2. 主题与消息模型

### 2.1 主题（Topic）
- 主消息：`chat.messages`
- 死信队列：`chat.messages.dlq`

### 2.2 事件结构
`MessageEvent` 结构：
- `ID`：消息 ID（用于幂等）
- `SessionID`：会话 ID（用于分区键，保证同一会话内顺序）
- `Role` / `Content` / `CreatedAt`

## 3. 写路径策略（Producer）

### 3.1 同步与确认
- 默认使用 **同步发送**（`async=false`）
- `required_acks=all`：需 ISR 全部确认后才认为写入成功

作用：
- 避免 Leader 宕机导致已返回成功但实际上消息未复制的风险。

### 3.2 批量与压缩
- `batch_size` + `batch_timeout_ms` 控制批量发送
- `compression=snappy` 降低带宽和磁盘开销

作用：
- 高并发时提升吞吐，减少网络瓶颈。

### 3.3 失败可见与重试
- `max_attempts`、`write_timeout_ms` 控制重试与超时
- 业务层对发送失败记录日志（不吞掉错误）

## 4. 读路径策略（Consumer）

### 4.1 手动提交 Offset
Consumer 使用 `FetchMessage` + `CommitMessages`：
- **仅当处理成功后提交**
- 避免自动提交导致的“已提交但未处理”消息丢失

### 4.2 有限重试 + 退避
- `max_retry` 控制最大重试次数
- `retry_backoff_ms` 退避时间

作用：
- 降低瞬时故障（例如数据库短暂不可用）导致的消息丢失。

### 4.3 死信队列（DLQ）
重试失败后写入 `chat.messages.dlq`：
- 保留原始 Key/Value
- 附加 `dlq_error` Header

作用：
- 保障异常消息可追踪、可补偿，避免阻塞主消费链路。

## 5. 幂等与一致性策略

### 5.1 MySQL 幂等
消息落库依赖主键：
- `MessagePersistHandler` 对重复消息创建时忽略主键冲突

作用：
- 避免 Consumer 重启或重复消费导致重复写入。

### 5.2 顺序保证
- 使用 `SessionID` 作为消息 Key
- Kafka 同一分区内有序，保证同会话消息顺序

## 6. 高并发与吞吐策略
- 批量写入 + 压缩减少吞吐压力
- Min/Max Bytes 与 MaxWait 提升拉取效率
- Consumer group 支持水平扩展（同组多个实例）

## 7. 高可用与容错

### 7.1 生产部署
`deploy/docker-compose.prod.yml` 中配置 3 Broker：
- `KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=3`
- `KAFKA_MIN_INSYNC_REPLICAS=2`
- `KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR=3`
- `KAFKA_TRANSACTION_STATE_LOG_MIN_ISR=2`

作用：
- 单节点故障仍可保证消息写入和消费。

### 7.2 当前限制
- Zookeeper 在 compose 中仍为单点（可用性瓶颈）。
- 生产建议升级为 3 ZK 或 KRaft。

## 8. 关键配置清单

```yaml
kafka:
  brokers:
    - "localhost:9092"
  dlq_topic: "chat.messages.dlq"
  producer:
    required_acks: "all"
    async: false
    batch_size: 100
    batch_timeout_ms: 10
    compression: "snappy"
    write_timeout_ms: 10000
    max_attempts: 10
  consumer:
    group_id: "coca-chat-consumer"
    min_bytes: 10000
    max_bytes: 10485760
    max_wait_ms: 500
    start_offset: "latest"
    max_retry: 3
    retry_backoff_ms: 200
    commit_timeout_ms: 3000
```

## 9. 故障场景与应对

### 9.1 Kafka Broker 宕机
- 多副本 + ISR 保障写入可继续
- Producer 使用 `acks=all`，避免写入丢失

### 9.2 Consumer 宕机
- 未提交 offset 的消息会被重新消费
- 幂等写入避免重复落库

### 9.3 MySQL 故障
- 消息在 Kafka 中持久化堆积
- MySQL 恢复后可继续消费

### 9.4 异常消息
- 重试失败后进入 DLQ
- 不阻塞主链路

## 10. 已知限制与后续优化
- DLQ 目前只记录错误 Header，建议增加监控与告警。
- Producer/Consumer 还未引入事务性 exactly-once 语义。
- Zookeeper 单点影响生产稳定性，建议升级为集群。

----

这套设计保证了在高并发下消息写入快速响应，同时通过手动提交、有限重试、DLQ 与幂等写入保障可靠性与容错能力。
