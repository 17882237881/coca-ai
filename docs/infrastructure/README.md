# 基础设施快速开始

## 目录结构
- `docker-compose.yml`: 主编排文件。
- `infra/data/`: 自动生成的数据持久化目录。**请勿手动修改**。
- `config/`: 配置文件 (如 Prometheus)。

## 如何启动

1. **启动所有服务**:
   ```powershell
   docker-compose up -d
   ```

2. **验证容器状态**:
   ```powershell
   docker-compose ps
   ```
   确保所有容器处于 `Up` 状态。

## 服务访问看板

| 组件 | 宿主机端口 | 内部 Hostname | 凭证 (如有) |
| :--- | :--- | :--- | :--- |
| **MySQL** | `3306` | `coca-mysql` | `root` / `root` |
| **Redis** | `6379` | `coca-redis` | 无 |
| **Milvus** | `19530` | `coca-milvus` | 默认 |
| **MinIO (Console)** | `9001` | `coca-minio` | `minioadmin` / `minioadmin` |
| **Kafka (Ext)** | `9094` | `coca-kafka` | 无 |
| **Jaeger UI** | `16686` | `coca-jaeger` | 无 |
| **Prometheus** | `9090` | `coca-prometheus` | 无 |

## 故障排除
- 如果 Milvus 启动失败，请先检查 `etcd` 和 `minio` 是否健康。
- 如果 Kafka 启动失败，请确保 `zookeeper` 已启动。
