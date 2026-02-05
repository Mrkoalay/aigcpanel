# AIGCPanel Golang 全量重写方案（接口 + 入库）

> 目标：将当前项目的核心后端能力统一到 Go 服务中，覆盖 API、数据库持久化、任务调度与模型调用编排。

## 当前已落地（本次代码）

- Go 服务入口：`cmd/aigcpaneld`
- 数据库：文件入库（JSON File DB）
- 数据初始化：`users / servers / tasks` 三类核心数据自动初始化
- API：
  - `GET /health`
  - `GET/POST /api/v1/users`
  - `GET/POST /api/v1/servers`
  - `PATCH /api/v1/servers/{id}`
  - `GET/POST /api/v1/tasks`
  - `PATCH /api/v1/tasks/{id}`
  - `POST /api/app_manager/user_info`（兼容现有 Electron 用户信息调用）
- 架构分层：`HTTP Handler -> Repository -> DB`

## 继续全量重写的实施路线

### 1. 接口全覆盖

- 对齐现有 Electron 侧调用协议（参数结构、错误码、流式输出）
- 增加认证、模型包管理、日志查询、统计等接口

### 2. 入库逻辑全覆盖

- 增加模型、任务日志、文件资源、版本升级记录等表
- 引入事务与乐观锁（关键状态流转）
- 对高频查询建立索引

### 3. 模型调用编排

- Go 作为控制平面，统一管理 Python 推理进程生命周期
- 支持任务队列、重试、失败补偿、取消任务

### 4. 可观测与稳定性

- 结构化日志 + TraceID
- 指标（QPS、任务耗时、错误率）
- 健康检查与依赖自检（DB、模型进程）

## 建议

- 第一阶段先确保 API 与数据库 schema 稳定，再迁移模型编排逻辑。
- 当前 JSON 文件库适合单机桌面场景，后续可平滑替换为 SQLite/PostgreSQL/MySQL。


## 兼容层策略

- 先提供 Electron 现有高频接口兼容路径（如 `/api/app_manager/user_info`），降低前端改造风险。
- 再逐步把 Electron `mapi` 逻辑迁移为调用 `/api/v1/*` 标准资源接口。
