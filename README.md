# Ginny Monorepo

Ginny 框架及生态组件 v2 统一管理。

## Framework

| 模块 | 说明 | 状态 |
|------|------|------|
| **[ginny](ginny/)** | 核心框架 — ConnectRPC + gRPC + gRPC-Web | ✅ v2 |
| [ginny-cli](ginny-cli/) | CLI 脚手架 | ✅ v2 |
| [ginny-demo](ginny-demo/) | 示例项目 | ✅ v2 |
| [ginny-template](ginny-template/) | 项目模板 | ✅ v2 |
| [ginny-component-template](ginny-component-template/) | 组件模板 | ✅ v2 |

## Components (v2)

| 组件 | 说明 |
|------|------|
| [ginny-broker](ginny-broker/) | 消息代理 (Kafka 集成) |
| [ginny-consul](ginny-consul/) | Consul 服务发现 |
| [ginny-encrypt](ginny-encrypt/) | AES 加密工具 |
| [ginny-desensitivity](ginny-desensitivity/) | 数据脱敏 |
| [ginny-gorm](ginny-gorm/) | GORM 数据库封装 |
| [ginny-mongo](ginny-mongo/) | MongoDB 封装 |
| [ginny-mysql](ginny-mysql/) | MySQL 封装 |
| [ginny-redis](ginny-redis/) | Redis 封装 |
| [ginny-locker](ginny-locker/) | 分布式锁 |
| [ginny-asynq](ginny-asynq/) | Asynq 任务队列 |
| [ginny-util](ginny-util/) | 工具库 (retry / snowflake / ip) |

## Removed (v2 核心已内置)

| 组件 | 替代方案 |
|------|---------|
| ~~ginny-util/graceful~~ | `errgroup` + LifecycleHook |
| ~~ginny-util/validation~~ | `interceptor/validation/` |

## 使用

```bash
cd ginny-monorepo
go work sync
```

https://github.com/richenlin/ginny-monorepo
