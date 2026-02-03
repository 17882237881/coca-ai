# Redis 设计说明（Coca-AI）

## 1. 角色与目标
Redis 在本项目中承担两类核心职责：
- **会话消息热缓存**：降低 MySQL 压力，提升高并发下的读取响应速度。
- **登录态黑名单**：用于用户登出后 SSRID 的失效校验（安全性优先，可配置降级）。

设计目标：
- **高并发读写**：消息读写在毫秒级；避免数据库成为瓶颈。
- **可控内存**：对会话消息缓存做 TTL 和长度限制，防止无限增长。
- **可用性与安全可切换**：登录态校验支持 Fail-Secure 或 Fail-Open。

相关实现位置：
- 初始化：`internal/ioc/redis.go`
- 消息缓存：`internal/repository/cache/message.go`
- 消息仓库：`internal/repository/message.go`
- 登录态校验：`internal/handler/middleware/login_jwt.go`
- 配置：`internal/config/config.go`, `configs/config.yaml`

## 2. 数据模型与 Key 设计

### 2.1 会话消息缓存
- Key 模式：`chat:session:{session_id}:messages`
- 数据结构：Redis List
- Value：`CachedMessage` 的 JSON 序列化
- TTL：24 小时（`messageTTL`）
- 最大长度：可配置（`redis.message_cache_max_len`）

列表结构用于保证同一会话内消息顺序，并以 `RPUSH` 追加。

### 2.2 登录态黑名单
- Key 模式：`users:ssid:{ssid}`
- Value：空字符串
- TTL：由登出逻辑传入（通常与 token 失效时间一致）

### 2.3 设计为什么
- List 支持高效追加和范围读取（`RPUSH` + `LRANGE`），适合按时间顺序存储会话消息。
- TTL 确保缓存自动清理，避免历史消息无限增长。
- LTRIM 控制最大长度，避免单会话撑爆 Redis 内存。

## 3. 读写路径与策略

### 3.1 写路径（Write-Behind）
1. 用户消息或 AI 消息生成后先写入 Redis（热缓存）。
2. 同时发送 Kafka，后续异步落库到 MySQL。

### 3.2 读路径（Read-Through）
1. 优先从 Redis 读取会话消息。
2. 若缓存未命中，回源 MySQL，并异步回填 Redis。

相关逻辑：`internal/repository/message.go`

## 4. 高并发与性能策略

### 4.1 Pipeline + 原子批处理
缓存写入采用 `TxPipeline`：
- `RPUSH` + `EXPIRE` + `LTRIM` 一起提交
- 减少 RTT，并保证写入和过期/裁剪行为一致

实现：`internal/repository/cache/message.go`

### 4.2 热数据拆分
每个会话独立一个 List，天然分散热点，避免全局锁。

### 4.3 缓存长度限制
通过 `LTRIM` 保留最近 N 条，避免内存膨胀。
- 配置项：`redis.message_cache_max_len`

### 4.4 连接池与超时
Redis 客户端支持：
- `pool_size`、`min_idle_conns`
- `dial_timeout_ms`、`read_timeout_ms`、`write_timeout_ms`

配置位置：`configs/config.yaml`

## 5. 可用性与容错策略

### 5.1 缓存写入失败不阻塞主流程
消息写入 Redis 失败时记录日志，但不阻塞主流程（避免单点影响核心业务）。

### 5.2 登录态校验 Fail-Secure / Fail-Open
- 默认 Fail-Secure：Redis 不可用时拒绝请求（安全优先）。
- 可配置 Fail-Open：Redis 故障时允许请求通过（可用性优先）。

配置项：`redis.fail_open`

### 5.3 过期与冷数据回落
即使 Redis 过期或丢失，消息仍可从 MySQL 回源（高可用保证）。

## 6. 一致性与风险控制

### 6.1 缓存一致性
- 写入走 Redis + Kafka，数据库由 Consumer 异步持久化。
- Redis 与 MySQL 可能短暂不一致（符合最终一致性模型）。

### 6.2 失效控制
- TTL + LTRIM 防止缓存无界增长。
- Redis 故障不影响消息最终落库（Kafka 保障）。

## 7. 部署与可用性形态

### 7.1 当前实现
- Redis 单实例（`docker-compose.yml` / `deploy/docker-compose.prod.yml`）。
- 持久化配置：`appendonly yes`。

### 7.2 建议生产形态
- Redis Sentinel 或 Cluster，实现高可用与自动故障转移。

## 8. 监控与运维建议
建议监控指标：
- 命中率（LRANGE 命中 vs MySQL 回源）
- 内存使用率 / key 数量
- 延迟 / 超时 / 连接池使用率

## 9. 关键配置清单

```yaml
redis:
  addr: "localhost:6379"
  password: ""
  db: 0
  pool_size: 50
  min_idle_conns: 10
  dial_timeout_ms: 3000
  read_timeout_ms: 3000
  write_timeout_ms: 3000
  message_cache_max_len: 200
  fail_open: false
```

## 10. 已知限制与后续优化
- Redis 单点：当前 compose 仍是单实例。
- 缓存一致性为最终一致性，若需强一致需同步落库或事务化补偿。
- 黑名单查询完全依赖 Redis，Redis 故障会影响鉴权路径。

----

这份设计强调 Redis 在高并发下的读写优化与可控内存，同时将安全性与可用性做成可配置策略，以适配生产需求。
