# 04. Kafka 运维与监控实战

## 1. 关键监控指标

运维 Kafka 时，除了看“死没死”，最重要的是看“堵没堵”。

### 1.1 Consumer Lag (消费积压)
这是 P0 级指标。
- **公式**: `Lag = LogEndOffset - CurrentOffset`
- **含义**: 还有多少条消息没处理。
- **报警阈值**: 建议设置为 `1000` 或 `5000`。
- **处理方案**: 
    1. 临时: 增加 Consumer 线程数/实例数。
    2. 长期: 优化 Consumer 里的 SQL 写入速度（比如改为批量插入）。

### 1.2 ISR Shrink / Expansion (ISR 抖动)
- **含义**: Partition 的副本同步队形乱了。
- **危险**: 如果频繁抖动，说明网络不稳定或磁盘 IO 瓶颈，可能导致数据丢失或 Leader 频繁切换（导致服务不可用）。

### 1.3 Active Controller Count
- **含义**: 正确值应该是 `1`。
- **异常**: 如果不是 1，说明集群脑裂 (Split Brain)，这是灾难性的。

## 2. 数据保留策略 (Retention)

磁盘不是无限的，Kafka 必须删除旧数据。Coca AI 的配置建议：

### 基于时间
`log.retention.hours=168` (7天)
- 超过 7 天的消息会被物理删除。

### 基于大小
`log.retention.bytes=1073741824` (1GB)
- 如果某个 Partition 超过 1GB，删掉最旧的 Segment。

### 针对 Coca AI 的特殊配置
由于我们的聊天记录最终去了 MySQL，Kafka 里的数据其实是“临时中转”。
- **建议**: 生产环境可以设置得短一点，比如 `24小时`。
- **Debug**: 开发环境保留长一点，方便回溯问题。

## 3. 灾难恢复 (Disaster Recovery)

### 场景：某条消息导致 Consumer 必定 Crash (Poison Pill)
有时候，Producer 发了一条畸形的 JSON，导致 Consumer 解析失败 Panic，或者触发了代码 Bug。
结果：Consumer 挂掉 -> 重启 -> 读到同一条毒药消息 -> 挂掉。死循环。

**解决方案**:
1. **Recover 中间件**: 在 Handler 外层加 `defer recover()`。
2. **Dead Letter Queue (DLQ, 死信队列)**:
    - 捕获处理失败的消息。
    - 把它转发到另一个 Topic `chat.messages.dlq`。
    - 提交原 Topic 的 Offset（跳过它）。
    - 既然无法处理，先隔离起来，人工后续排查，不要阻塞主流程。
