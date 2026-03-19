# codeScope Bridge

`bridge` 是 codeScope 的独立旁路采集器。

它不会启动、托管、包装或替换 `Codex CLI` / `Claude Code`。用户继续按原有方式独立运行这些工具，`bridge` 只负责：

- 连接 `server` 的 `/ws/bridge`
- 自动发现本机活跃的 `codex` / `claude` 会话
- 通过旁路 capture adapters 采集并标准化兼容事件
- 在可能的情况下补充 `project / thread / message / thread_state` 语义

## 当前真实状态

### 已实现

- 默认 `capture_mode=discovery`
- 进程发现与 workspace 识别
- `ProcessSnapshotAdapter`
- `SessionHeartbeatAdapter`
- `WorkspaceFileWatcherAdapter`
- Codex 本地 session 文件旁路读取
- 向 server 上报兼容层 `event` / `heartbeat` / `command_result`
- 在 payload 中补充部分 `project_name / workspace_root / thread_id / source_session_id / thread_title / thread_state`

### 半实现

- 真实消息采集目前主要覆盖 Codex 本地 `session_index.jsonl` 和 `sessions/*.jsonl`
- 线程语义已经开始脱离单纯 `session_id`，但还不是完整 thread runtime
- `thread_state` 仍然带启发式成分

### 未实现

- Claude Code 对等的真实消息采集适配器
- 完整 `mobile -> prompt -> bridge -> local agent -> message 回流` 闭环
- 默认 `bridge <-> mobile` P2P 内容通道

## 默认运行链路

1. `bridge` 启动并连接 `server`
2. 周期扫描本机进程，识别候选 `codex` / `claude` 会话
3. 为每个会话启动最小旁路采集链
4. 同时旁路读取可用的 Codex 本地 session 文件
5. 向 `server` 上报兼容事件和已派生出的高层语义字段

## 当前内置采集层

### `ProcessSnapshotAdapter`

- 上报一个 `command` 事件，包含 `pid`、`process_name`、`command_line`
- 上报一个 `terminal_output` 观察提示，表明该会话已被旁路发现

### `SessionHeartbeatAdapter`

- 为活跃会话周期性上报 `heartbeat`

### `WorkspaceFileWatcherAdapter`

- 基于 `workspace_root` 监听文件变化并上报 `file_change`
- 默认忽略 `.git`、`.codescope`、`node_modules`、`.dart_tool`、`build`、`dist`、`coverage`、`tmp`、`target`、`out`
- 对短时间重复文件变更做基础去重

### `CodexSessionSource`

- 读取 `~/.codex/session_index.jsonl`
- 读取最近的 `~/.codex/sessions/**/*.jsonl`
- 识别 `user_message` / `agent_message`
- 生成稳定的语义 `thread_id`
- 发出部分真实 `user / assistant` 内容

## 重要约束

- `terminal_output` 在默认 discovery 模式下是“bridge 观察提示”，不是 PTY 直连 stdout
- `heartbeat`、原始命令行、observing 提示都属于 debug/diagnose 层，不属于 release 主视图
- discovery 会尽量折叠同进程树、同 agent、同 workspace 的重复候选，减少伪 session
- discovery 会在稳定窗口内复用原 `session_id`，降低短暂抖动带来的 session 抖动

## 当前兼容协议

### bridge 上行

- `message_type=event`
- `message_type=heartbeat`
- `message_type=command_result`

### server 下行

- `message_type=command`
- `command_type=send_prompt`

说明：

- 这是兼容层协议
- 默认 side-channel mode 下，`send_prompt` 仍不假装成功注入独立 CLI

## 配置

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `CODESCOPE_BRIDGE_AGENT_NAME` | `bridge` | bridge 自身标识 |
| `CODESCOPE_BRIDGE_SERVER_URL` | `ws://localhost:8080/ws/bridge` | server WebSocket 地址 |
| `CODESCOPE_BRIDGE_WORKSPACE_ROOT` | `.` | 发现不到 workspace 时的回退根目录 |
| `CODESCOPE_BRIDGE_MACHINE_ID` | 自动生成 | 机器标识 |
| `CODESCOPE_BRIDGE_SESSION_ID` | 自动生成 | bridge 自身兼容 session 标识 |
| `CODESCOPE_BRIDGE_CAPTURE_MODE` | `discovery` | 默认旁路发现模式 |
| `CODESCOPE_BRIDGE_DISCOVERY_INTERVAL` | `5s` | 进程发现轮询间隔 |
| `CODESCOPE_BRIDGE_SESSION_HEARTBEAT_INTERVAL` | `15s` | 每个活跃会话 heartbeat 间隔 |
| `CODESCOPE_BRIDGE_SESSION_STABILITY_WINDOW` | `10s` | session 抖动复用窗口 |

辅助/测试模式仍保留 `reader` / `jsonl`、旧 prompt inbox 和旧 managed process 配置，但都不是当前默认主链路。

## 运行

### 默认旁路模式

```powershell
$env:CODESCOPE_BRIDGE_SERVER_URL="ws://localhost:8080/ws/bridge"
$env:CODESCOPE_BRIDGE_WORKSPACE_ROOT="D:\Work\Code\Cross\codeScope"
go run ./cmd/bridge
```

### 输入回放模式

```powershell
$env:CODESCOPE_BRIDGE_CAPTURE_MODE="reader"
$env:CODESCOPE_BRIDGE_SOURCE_MODE="jsonl"
go run ./cmd/fake-source -workspace-root "D:/Work/Code/Cross/codeScope" | go run ./cmd/bridge
```

## 验证

```powershell
go test ./...
go build ./...
```

## 旧实现如何理解

以下实现仍保留在代码库，但已降级为 legacy / extension points，不再是默认路径：

- `internal/capture/managed_process.go`
- `internal/command/consumer.go`
- `internal/command/FilePromptSink`
- `internal/command/ExecutionPromptSink`

它们当前不参与默认的 `bridge -> discovery -> side-channel capture -> server` 主链路。
