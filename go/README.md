# AIGCPanel Go Rewrite (Phase 1)

该目录提供了一个 **Golang 重写版本**，用于替代原项目的核心后端能力。

## 已实现能力

- 纯 Go（标准库）HTTP 服务
- JSON 文件持久化（自动建库）
- 用户、模型服务、语音配置、视频模板、任务的基础 CRUD（含状态更新）
- 清晰分层：`api` / `app` / `store` / `domain`

## 运行

```bash
cd go
go run ./cmd/aigcpaneld
```

默认监听 `:8080`，默认数据文件 `data/aigcpanel.json`。

### 环境变量

- `AIGCPANEL_ADDR`：监听地址
- `AIGCPANEL_DSN`：JSON 数据文件路径

## 接口

- `GET /health`
- `GET/POST /api/v1/users`
- `GET/POST /api/v1/servers`
- `PATCH /api/v1/servers/{id}`
- `GET/POST /api/v1/voice-profiles`
- `GET/POST /api/v1/video-templates`
- `GET/POST /api/v1/tasks`
- `PATCH /api/v1/tasks/{id}`

## 测试

```bash
cd go
go test ./...
```
