# Ginny Monorepo

Ginny 生态所有组件的统一管理仓库。

## Modules

| 组件 | 说明 |
|------|------|
| ginny-asynq | Asynq 任务队列集成 |
| ginny-broker | 消息代理 |
| ginny-cli | CLI 脚手架工具 |
| ginny-component-template | 组件模板 |
| ginny-config | 配置管理组件 |
| ginny-consul | Consul 服务发现 |
| ginny-demo | 示例项目 |
| ginny-desensitivity | 数据脱敏 |
| ginny-encrypt | 加密工具 |
| ginny-gorm | GORM 数据库集成 |
| ginny-jaeger | Jaeger 链路追踪 |
| ginny-locker | 分布式锁 |
| ginny-log | 日志组件 |
| ginny-mongo | MongoDB 集成 |
| ginny-mysql | MySQL 集成 |
| ginny-prometheus | Prometheus 监控 |
| ginny-redis | Redis 集成 |
| ginny-serve | 服务管理 |
| ginny-template | 项目模板 |
| ginny-util | 工具库 |

## 使用

本仓库使用 Go workspace (`go.work`) 管理多模块：

```bash
cd ginny-monorepo
go work sync
```

每个组件保留独立 `go.mod` 和版本历史（`git subtree`）。
