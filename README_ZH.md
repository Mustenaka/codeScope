# codeScope

[English README](./README.md)

codeScope 是一个让你暂时离开电脑后，仍然能在手机上查看 AI coding agent 正在做什么、并继续同一条工作线程的跨设备系统。

它的目标产品模型是：

```text
device -> project -> thread -> message -> thread_state -> file browser
```

也就是说，主视图不应该是底层 `session/event` 调试面板，而应该让用户直接看到：

- 这台电脑上有哪些项目正在活跃
- 每个项目里有哪些 agent 对话线程
- 用户发送了什么，agent 回复了什么
- 每个线程当前是运行中、等待用户、已完成、阻塞、陈旧还是离线
- 哪些项目文件可以在手机上查看

## 仓库结构

- `bridge/`：本地旁路采集器，发现运行中的 agent 进程并上报兼容事件与部分高层语义
- `server/`：当前的 REST / WebSocket 后端，以及派生出的 `project/thread/message` 读模型
- `mobile/`：Flutter 客户端，负责项目、线程、消息、文件浏览和 prompt 交互
- `doc/`：核心产品、协议、路线图和维护文档
- `docs/`：开发过程中产生的辅助实现笔记

## 当前状态

### 已实现

- `bridge` 以 side-channel `discovery` 模式运行，不改变用户原有的 Codex / Claude 启动方式。
- `bridge` 可以发现本机 `codex` / `claude` 进程、监听工作区文件变化，并上报兼容层事件。
- `bridge` 还能读取近期 Codex 本地 session 文件，补充部分真实 `user / assistant` 内容。
- `server` 已能从当前 ingestion 模型中派生 `project`、`thread`、`message` 只读 API，同时保留 legacy `sessions` 接口。
- `mobile` 主导航已经切到 `project -> thread -> message`。

### 半实现

- `thread` 身份在一些链路里仍然带有 legacy `session` 痕迹。
- `thread_state` 在部分路径里仍依赖启发式推导。
- 真实消息采集目前对 Codex 支持更强，对其他 agent 还不完整。
- 列表实时刷新仍然混合使用增量更新和快照重拉。

### 未实现

- 所有受支持 agent 的完整真实输入/输出采集
- 项目级 `create thread` 主链路
- 正式的 `bridge <-> mobile` P2P 直连与 relay fallback
- 严格可信的 thread 生命周期状态机

## 架构方向

codeScope 当前遵循这些原则：

- `bridge` 保持为旁路采集器，而不是托管启动器
- `server` 默认应收敛到发现、路由和可选 relay fallback，而不是永久内容中心
- 消息正文、线程历史、文件内容最终应优先走 `bridge <-> mobile` 直连
- heartbeat、原始命令行、bridge observing 提示、底层 payload 只属于 debug 视图，不属于 release 主视图

## 快速开始

### 启动 server

```powershell
cd server
go run ./cmd/server
```

### 启动 bridge

```powershell
cd bridge
go run ./cmd/bridge
```

### 不依赖 mobile 的兼容层联调

```powershell
cd bridge
$env:CODESCOPE_BRIDGE_SERVER_URL="ws://localhost:8080/ws/bridge"
$env:CODESCOPE_BRIDGE_SESSION_ID="session-demo"
$env:CODESCOPE_BRIDGE_SOURCE_MODE="jsonl"
go run ./cmd/fake-source -workspace-root "D:/Work/Code/Cross/codeScope" | go run ./cmd/bridge
```

### 启动 debug subscriber

```powershell
cd server
go run ./cmd/debug-subscriber -server-url ws://localhost:8080/ws/mobile -session-id session-demo
```

### 检查接口

```powershell
curl http://localhost:8080/api/health
curl http://localhost:8080/api/sessions
curl http://localhost:8080/api/projects
```

## 验证

请分别在各模块目录内执行：

```powershell
cd bridge
go test ./...
go build ./...

cd ../server
go test ./...
go build ./...

cd ../mobile
flutter test
```

## 文档入口

建议按以下顺序阅读：

1. [`doc/全局修改指南-AI-Agent协作基线.md`](./doc/%E5%85%A8%E5%B1%80%E4%BF%AE%E6%94%B9%E6%8C%87%E5%8D%97-AI-Agent%E5%8D%8F%E4%BD%9C%E5%9F%BA%E7%BA%BF.md)
2. [`doc/重构方案-项目线程消息与P2P直连.md`](./doc/%E9%87%8D%E6%9E%84%E6%96%B9%E6%A1%88-%E9%A1%B9%E7%9B%AE%E7%BA%BF%E7%A8%8B%E6%B6%88%E6%81%AF%E4%B8%8EP2P%E7%9B%B4%E8%BF%9E.md)
3. [`doc/bridge/Bridge旁路采集需求重设计.md`](./doc/bridge/Bridge%E6%97%81%E8%B7%AF%E9%87%87%E9%9B%86%E9%9C%80%E6%B1%82%E9%87%8D%E8%AE%BE%E8%AE%A1.md)
4. [`doc/bridge-server-完整协议草案.md`](./doc/bridge-server-%E5%AE%8C%E6%95%B4%E5%8D%8F%E8%AE%AE%E8%8D%89%E6%A1%88.md)
5. [`doc/正式版本落地路线图.md`](./doc/%E6%AD%A3%E5%BC%8F%E7%89%88%E6%9C%AC%E8%90%BD%E5%9C%B0%E8%B7%AF%E7%BA%BF%E5%9B%BE.md)
6. [`doc/执行提示词-正式版落地.md`](./doc/%E6%89%A7%E8%A1%8C%E6%8F%90%E7%A4%BA%E8%AF%8D-%E6%AD%A3%E5%BC%8F%E7%89%88%E8%90%BD%E5%9C%B0.md)

## 下一步重点

1. 补齐真实消息采集，减少对 session 派生 thread 的依赖。
2. 把 `thread_state` 收敛成正式生命周期状态机。
3. 完成手机端继续发 prompt 的正式闭环。
4. 落地 `bridge <-> mobile` P2P 直连和 relay fallback。
5. 把项目级 file browser 和 create-thread 体验补完整。
