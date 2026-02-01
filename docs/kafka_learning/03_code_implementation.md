# 03. Coca AI 代码实现详解：进阶篇

## 1. Producer 的高级配置

在 `internal/mq/producer.go` 中，我们不仅是简单的 NewWriter，还涉及了性能调优的关键参数。

```go
writer := &kafka.Writer{
    Addr:     kafka.TCP(cfg.Brokers...),
    Topic:    TopicChatMessages,
    Balancer: &kafka.KeyHash{}, 
    
    // --- 性能调优参数 ---
    
    // 1. 批量发送
    // 默认是 1。设置为 100 意味着：
    // 攒够 100 条消息 OR 距离第一条消息过了 10ms，才会真正发起一次网络 TCP 请求。
    // 这将 QPS 提升了几个数量级。
    BatchSize:    100,
    BatchTimeout: 10 * time.Millisecond,
    
    // 2. 压缩
    // 文本数据 (JSON) 压缩率极高 (通常 1/10)。
    // 常用的有 Snappy (谷歌轻量级, 速度快) 或 Gzip (压缩率高, CPU 稍高)
    // 开启压缩能极大节省网卡带宽和 Kafka 磁盘空间。
    Compression: kafka.Snappy, 

    // 3. 异步非阻塞
    // 如果为 false，WriteMessages 会一直阻塞直到 Kafka 返回 ACK。
    // 如果为 true，WriteMessages 把消息放入内存 Buffer 就返回 nil，后台 Goroutine 负责发送。
    // 风险：如果进程突然 crash，内存 Buffer 里的消息会丢。对于聊天记录，我们选择 true 换取低延迟。
    Async: true,
}
```

## 2. Consumer 的优雅关闭与上下文控制

在 `internal/mq/consumer.go` 中，`Start` 方法只是冰山一角。在生产环境中，**优雅关闭 (Graceful Shutdown)** 至关重要。

### 问题场景
还在消费一条消息写数据库，这时候 Kubernetes/Docker 把 Pod 杀掉了。如果直接断电，消费者可能已经写了数据库，但还没来得及提交 Offset。重启后，这条消息又来了 -> **重复消费**。

### 完整实现代码 (伪代码)

```go
func (c *Consumer) Start(ctx context.Context) {
    // 注册系统信号监听 (Ctrl+C, kill)
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        <-sigChan
        log.Println("Shutting down consumer...")
        // 关键：关闭 Reader，这会让下面 FetchMessage 返回 io.EOF 或 error
        c.reader.Close()
    }()

    for {
        msg, err := c.reader.FetchMessage(ctx)
        if err != nil {
            if err == io.EOF {
                log.Println("Consumer closed.")
                return // 正常退出
            }
            log.Printf("Read error: %v, retrying...", err)
            time.Sleep(time.Second) // 错误退避，防止死循环刷日志
            continue
        }

        // 处理逻辑...
        process(msg)

        // 提交
        c.reader.CommitMessages(ctx, msg)
    }
}
```

## 3. Go Channel 模式 vs Handler 模式

在代码实现中，我们设计了 `RegisterHandler` 接口：

```go
type Handler func(ctx context.Context, event *MessageEvent) error

func (c *Consumer) RegisterHandler(h Handler) {
    c.handlers = append(c.handlers, h)
}
```

**设计思考**:
- 这是一种 **Observer 模式**。
- `Consumer` 结构体不需要知道具体的业务逻辑（是存 MySQL 还是发邮件），它只负责“搬运”。
- `cmd/server/wire.go` 或者 `internal/ioc/kafka.go` 负责把具体的 `MessagePersistHandler` 注入进去。
- **扩展性**: 如果明天因为法律要求，所有聊天记录要同时备份到 S3。你只需要写一个 `S3BackupHandler`，注册进去，不需要改动 Consumer 的任何核心代码。

## 4. 泛型消息处理 (Generics) - 未来优化方向

目前我们硬编码了 `MessageEvent`。如果系统复杂了，有 `UserLoginEvent`, `OrderPaidEvent` 怎么办？

可以使用 Go 1.18+ 的泛型：

```go
type KafkaConsumer[T any] struct {
    reader *kafka.Reader
    handlers []func(T)
}

func (c *KafkaConsumer[T]) Start() {
    // ...
    var event T
    json.Unmarshal(msg.Value, &event)
    // ...
}
```
这能进一步减少重复代码。
