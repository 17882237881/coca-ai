# 02. Coca AI 架构深度解析：为什么选 Kafka？

## 1. 竞品对比：Kafka vs RabbitMQ vs RocketMQ

在技术选型时，为什么我们为 Coca AI选择了 Kafka？

| 特性 | **Kafka** (本项目) | **RabbitMQ** | **RocketMQ** |
| :--- | :--- | :--- | :--- |
| **设计初衷** | 大数据日志处理，高吞吐 | 传统金融/任务队列，低延迟 | 阿里电商业务，高可靠+高吞吐 |
| **吞吐量** | **百万级/秒 (最高)** | 万级/秒 | 十万级/秒 |
| **持久化** | 磁盘文件 (Segment)，天然持久 | 内存为主，磁盘为辅 | 磁盘文件 (CommitLog) |
| **消息堆积** | **极强 (支持 TB 级堆积)** | 弱 (堆积多了性能骤降) | 强 |
| **消费模式** | Pull (消费者主动拉) | Push (Broker 推送) | Pull / Push |
| **适用场景** | **日志、流计算、用户行为追踪** | 订单处理、即时任务 | 交易支付、复杂业务路由 |

**Coca AI 选型理由**:
1.  **海量对话日志**: AI 对话数据类似 Log，未来可能用于训练 RAG 或 Fine-tuning，不仅是“发完就删”，更需要“持久保存”以便后续批量读取。
2.  **削峰能力**: 大语言模型 (LLM) 的响应具有突发性。Kafka 的磁盘顺序写特性让它能承受极高的写入并发。
3.  **生态**: 配合 Go 和 Python (未来) 做数据管道，Kafka 是标准。

## 2. 深入 Write-Behind (异步写后) 模式

我们在 Coca AI 中实现的不仅是简单的“异步”，而是一个完整的 **Write-Through Cache + Write-Behind Store** 模式。

### 异常场景推演

#### 场景 1：MySQL 突然宕机
-   **现象**: `MessagePersistHandler` (Consumer) 尝试执行 `INSERT` 失败。
-   **处理**: 
    -   Consumer 会捕获错误。
    -   **关键**: Consumer **不提交 Offset** (NACK)。
    -   Kafka 认为这条消息“没被处理成功”。
    -   Consumer 稍后会再次读取到这条消息 (重试)。
-   **用户感知**: 用户完全无感知。因为用户的 `GET /chat/history` 读的是 Redis 缓存，或者因为 MySQL 挂了暂时读不到旧历史，但**新发的消息不会丢**，它们安全地积压在 Kafka 磁盘上。
-   **恢复**: MySQL 重启后，Consumer 快速消费积压数据，数据最终一致。

#### 场景 2：Redis 满了 / 挂了
-   **现象**: API 写入 Redis 失败。
-   **处理**: 
    -   代码中捕获 Redis 错误，记录 Log，但 **不阻断流程**。
    -   继续发送 Kafka。
-   **用户感知**: 可能会发现“刚发的消息刷新一下看不到了”（因为缓存没写进，MySQL 还没来得及写进）。
-   **最终一致**: 等 Consumer 把数据写入 MySQL 后，用户再次刷新，从 MySQL 读到了数据（Cache Miss -> Load from DB），恢复正常。

## 3. 幂等性设计 (Idempotency)

在分布式系统中，**Exactly-Once (精确一次)** 很难，通常我们保证 **At-Least-Once (至少一次)**。这意味着消息可能会**重复**。

**Coca AI 为什么要防重？**
如果不防重，Consumer 如果发生网络抖动（消息处理了但没提交 Offset），会重发消息。如果不处理，聊天记录里会出现两条这一样的“你好”。

**解决方案**:
1.  **数据库唯一索引 (Hard Check)**: 
    -   MySQL 表结构中，`message_id` 是 Primary Key 或 Unique Key。
    -   `INSERT IGNORE INTO messages` 或者 `ON DUPLICATE KEY UPDATE`。
2.  **业务层去重**:
    -   Consumer 可以在 Redis 里维护一个 `BitMap` 或 `Bloom Filter`，记录最近处理过的 MessageID。

本项目当前采用了 **数据库主键抗重** 的简单策略。
