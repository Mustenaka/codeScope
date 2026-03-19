# bridge-server 接口对接与测试
> 更新日期：2026-03-19
> 性质：兼容层联调文档

## 1. 文档定位

这份文档只描述当前代码里 `bridge <-> server` 的兼容层联调方式。

它不是新的产品协议总览。新的总方向请以：

- [全局修改指南-AI-Agent协作基线.md](/D:/Work/Code/Cross/codeScope/doc/全局修改指南-AI-Agent协作基线.md)
- [bridge-server-完整协议草案.md](/D:/Work/Code/Cross/codeScope/doc/bridge-server-完整协议草案.md)

为准。

## 2. 当前真实兼容状态

当前 `bridge` 与 `server` 之间仍然主要通过 legacy `session/event` 协议联调。

### 已兼容

- `bridge -> server` 的 `event`
- `bridge -> server` 的 `heartbeat`
- `bridge -> server` 的 `command_result`
- `server -> bridge` 的 `command(send_prompt)`

### 明确限制

- side-channel / discovery 模式下，`send_prompt` 当前会失败
- 失败原因是：`side-channel mode does not support prompt injection`
- 这属于当前能力边界，不是接口不兼容

## 3. 当前联调目标

当前文档的联调目标只应是：

- 验证 `bridge` 能稳定接入 `server`
- 验证兼容层 `session/event` 数据能被 ingest
- 验证 command task 和 command_result 闭环

不应把它误解成“已经完成项目/线程/消息主链路”。

## 4. 当前兼容协议

### bridge 接入

- `GET /ws/bridge`

### bridge 上行

- `message_type=event`
- `message_type=heartbeat`
- `message_type=command_result`

### server 下行

- `message_type=command`
- `command_type=send_prompt`

## 5. 当前联调步骤

### 步骤 1：启动 server

```powershell
cd server
go run ./cmd/server
```

### 步骤 2：启动 bridge

```powershell
cd bridge
$env:CODESCOPE_BRIDGE_SERVER_URL="ws://localhost:8080/ws/bridge"
$env:CODESCOPE_BRIDGE_CAPTURE_MODE="discovery"
$env:CODESCOPE_BRIDGE_WORKSPACE_ROOT="D:\Work\Code\Cross\codeScope"
go run ./cmd/bridge
```

### 步骤 3：确认兼容 session 已出现

```powershell
Invoke-RestMethod http://localhost:8080/api/sessions
```

期望：

- 出现对应 `session_id`
- `bridge_online = true`

### 步骤 4：检查兼容事件写入

```powershell
Invoke-RestMethod http://localhost:8080/api/sessions/<session_id>/events
```

期望：

- 能看到 `heartbeat`
- 能看到 discovery / file_change / terminal_output 等兼容事件

注意：

- 这些事件主要用于当前联调和诊断
- 不代表 release 产品主界面就应该显示这些内容

### 步骤 5：验证 command task 闭环

```powershell
$body = @{ content = "continue fixing tests" } | ConvertTo-Json
Invoke-RestMethod -Method Post -Uri http://localhost:8080/api/sessions/<session_id>/commands/prompt -ContentType "application/json" -Body $body
Invoke-RestMethod http://localhost:8080/api/sessions/<session_id>/commands
```

期望：

- command task 创建成功
- 最终状态为 `failed`
- 结果里出现 side-channel 不支持 prompt injection 的说明

### 步骤 6：验证离线状态

停止 `bridge` 后执行：

```powershell
Invoke-RestMethod http://localhost:8080/api/sessions/<session_id>
```

期望：

- `bridge_online = false`

## 6. 后续要迁移到什么

当前联调通过后，下一阶段不是继续堆 `sessions/events` 能力，而是迁移到：

- `device`
- `project`
- `thread`
- `message`
- `thread_state`
- `p2p route`

## 7. 结论

这份文档现在只用于维护“现有兼容链路可联调”。真正的产品和协议重心已经切换到：

- 首页看项目
- 项目里看线程
- 线程里看消息历史
- server 做发现和路由
- 内容优先 P2P
