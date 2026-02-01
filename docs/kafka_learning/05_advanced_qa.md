# 05. Kafka 高频面试题与实战 Q&A

这里整理了关于 Kafka 最常见的问题，既是面试必问，也是实战必坑。

## Q1: Kafka 如何保证消息顺序？
**A**: **Kafka 只能保证 Partition 内有序，不保证全局有序。**
- **实战**: 在 Coca AI 中，我们必须保证同一个 Session 的消息顺序（不能先回答再提问）。
- **做法**: Producer 发送时，将 `SessionID` 作为 `Key`。
- **原理**: Kafka 的默认 Partitioner 会对 Key 进行 Hash: `hash(SessionID) % partition_count`。这样，SessionID=1001 的所有消息永远会发给 Partition-0。Partition-0 会按顺序追加写入，Consumer 也是按顺序读取，从而保证局部有序。

## Q2: 消息重复消费了怎么办？(幂等性)
**A**:
- **原因**: 消费者处理完业务（写了 DB），但在提交 Offset 之前 crash 了。重启后 Kafka 不知道你处理过，又发了一遍。
- **解决**: 业务层必须幂等。
    - **DB**: 使用 `INSERT IGNORE` 或 `REPLACE INTO`。
    - **Redis**: 记录 `SET processed_msg_ids {id}`。

## Q3: 消息丢了怎么办？
**A**: 
1. **Producer 端**: 
    - `acks=0`: 丢了不知道。
    - `acks=1`: Leader 盘坏了会丢。
    - `acks=all`: 几乎不丢。
    - 代码层：检查 `WriteMessages` 返回的 error，报错必须重试。
2. **Consumer 端**:
    - **自动提交 (Auto Commit)** 的坑：如果开启自动提交，Kafka 可能会在每 5 秒自动提交一次。如果你拿到消息，还没处理完（比如存 DB 耗时长），Offset 已经被自动提交了。这时候如果你 crash 了，这条消息就**丢了**（Kafka 以为你处理完了）。
    - **解决**: Coca AI 关闭了自动提交（或者小心控制），我们在代码里明确执行 `CommitMessages`。**必须是“先处理，后提交”**。

## Q4: 怎么在这个项目里增加“敏感词过滤”功能？
**A**: 利用 Kafka 的解耦特性。
1. **方案 A (拦截器)**: 修改 Producer，发送前检查。缺点是阻塞用户响应。
2. **方案 B (独立服务)**: 
    - 新增一个 Consumer Group `audit-group`。
    - 异步读取 `chat.messages`。
    - 发现敏感词 -> 标记 MessageID -> 异步更新 MySQL 状态为“不可见” / 发送告警。
    - **优点**: 完全不影响主聊天的速度 (Zero Latency Impact)。

## Q5: Kafka 既然是写文件的，为什么还这么快？
**A**:
1. **PageCache**: Kafka 极度依赖操作系统的 PageCache（文件缓存）。大部分读写其实都是在内存里操作的，操作系统负责异步刷盘。
2. **Zero-Copy**: 零拷贝技术。
3. **Sequential I/O**: 顺序写盘，没有磁头寻道时间。
4. **Batching**: 批量操作，减少系统调用和网络开销。
