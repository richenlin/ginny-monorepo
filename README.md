# Ginny Monorepo

Ginny 框架及其生态组件的统一管理仓库。

## Framework | 框架

| 模块 | 说明 |
|------|------|
| **[ginny](ginny/)** | **核心框架** — Schema 驱动 Go RPC 框架 (ConnectRPC + v2) |

## Components | 组件

### 基础设施

| 组件 | 说明 |
|------|------|
| [ginny-config](ginny-config/) | 配置管理 |
| [ginny-log](ginny-log/) | 日志组件 |
| [ginny-util](ginny-util/) | 工具库 |
| [ginny-serve](ginny-serve/) | 服务管理 |

### 数据存储

| 组件 | 说明 |
|------|------|
| [ginny-gorm](ginny-gorm/) | GORM 数据库集成 |
| [ginny-mysql](ginny-mysql/) | MySQL 集成 |
| [ginny-mongo](ginny-mongo/) | MongoDB 集成 |
| [ginny-redis](ginny-redis/) | Redis 集成 |

### 中间件 & 可观测性

| 组件 | 说明 |
|------|------|
| [ginny-broker](ginny-broker/) | 消息代理 |
| [ginny-asynq](ginny-asynq/) | Asynq 任务队列 |
| [ginny-consul](ginny-consul/) | Consul 服务发现 |
| [ginny-jaeger](ginny-jaeger/) | Jaeger 链路追踪 |
| [ginny-prometheus](ginny-prometheus/) | Prometheus 监控 |
| [ginny-locker](ginny-locker/) | 分布式锁 |

### 安全

| 组件 | 说明 |
|------|------|
| [ginny-encrypt](ginny-encrypt/) | 加密工具 |
| [ginny-desensitivity](ginny-desensitivity/) | 数据脱敏 |

### 开发工具

| 组件 | 说明 |
|------|------|
| [ginny-cli](ginny-cli/) | CLI 脚手架 |
| [ginny-template](ginny-template/) | 项目模板 |
| [ginny-component-template](ginny-component-template/) | 组件模板 |
| [ginny-demo](ginny-demo/) | 示例项目 |

## 使用

Go workspace 管理所有模块：

```bash
cd ginny-monorepo
go work sync
```

## 仓库

https://github.com/richenlin/ginny-monorepo
