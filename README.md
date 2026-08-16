# 本地安全密码管理器 CLI

`cry013` 是面向本地演示和 Go 问题排查训练的完整可运行项目。业务服务使用 Go 1.24、Gin、GORM、validator、Zap 与 Viper；数据适配器为 SQLITE，界面形态为 无前端；提供版本化 HTTP API 或本地 CLI。

## 业务闭环

- AES-256-GCM 保险库
- Argon2id 密钥派生
- 原子备份与恢复
- 掩码检索和自动锁定

项目保留用户、角色、保险库、条目、活动、审计、刷新令牌、幂等写入和事务状态流转等公共能力。演示数据可重复初始化且不会覆盖用户数据。

## 目录

- `cmd/`：服务与 CLI 入口
- `internal/domain`：实体、权限和状态机
- `internal/application`：用例与事务边界
- `internal/repository`：内存及 sqlite 适配器
- `internal/transport/http`：`/api/v1` API
- `api/openapi`、`migrations`、`deploy`：接口、迁移与部署
- `web/`：前端（若项目需要）

## 验证

```bash
go build ./...
go test ./...
go test -race ./...
go vet ./...
```

服务默认使用内存演示仓储：`ALLOW_IN_MEMORY=true go run ./cmd/server`。容器启动使用 `docker compose up --build`。健康检查为 `/healthz` 与 `/readyz`。配置参考 `.env.example`；仓库不包含真实密钥、依赖缓存、构建产物或运行期数据库。
