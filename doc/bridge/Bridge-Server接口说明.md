# Bridge-Server 接口说明
> 更新日期：2026-03-19
> 性质：当前兼容接口说明，不代表最终产品协议

## 1. 文档定位

本文档描述的是当前代码里 `bridge` 与 `server` 之间仍在使用的兼容接口。

它的作用是：

- 支撑当前联调
- 解释 legacy `session/event` 协议
- 给迁移到新协议前的兼容实现提供边界

它不再定义最终产品主模型。最终方向请以：

- [全局修改指南-AI-Agent协作基线.md](/D:/Work/Code/Cross/codeScope/doc/全局修改指南-AI-Agent协作基线.md)
- [bridge-server-完整协议草案.md](/D:/Work/Code/Cross/codeScope/doc/bridge-server-完整协议草案.md)

为准。

## 2. 总原则

- `bridge` 是独立运行的旁路采集器
- `bridge` 不负责托管启动 `Codex CLI` / `Claude Code`
- `server` 不能依赖 `bridge` 的内部采集实现
- 当前兼容接口存在，但它不是未来 release 产品的主协议

## 3. 当前兼容接口

### 3.1 WebSocket

- `GET /ws/bridge`

### 3.2 bridge 上行消息类型

- `event`
- `heartbeat`
- `command_result`

### 3.3 server 下行消息类型

- `command`

### 3.4 当前支持的命令

- `command_type=send_prompt`

说明：

- 这是兼容层命令
- side-channel / discovery 下当前通常不会真正执行成功

## 4. 当前兼容消息结构

```json
{
  "message_id": "msg-123",
  "session_id": "session-1",
  "message_type": "event",
  "event_type": "terminal_output",
  "timestamp": "2026-03-19T10:00:00Z",
  "payload": {}
}
```

说明：

- `session_id` 仍然是当前兼容模型中的主键
- 后续会逐步迁移为 `device/project/thread`

## 5. 当前兼容事件类型

- `terminal_output`
- `ai_output`
- `command`
- `file_change`
- `heartbeat`
- `error`

注意：

- 这些类型是联调和诊断数据
- 不代表用户主视图应该直接看到这些类型

## 6. 当前 ACK / Error

server 仍会对 bridge 的兼容消息回 ACK：

```json
{
  "type": "ack",
  "message_id": "msg-123",
  "session_id": "session-1"
}
```

错误示例：

```json
{
  "type": "error",
  "error": "unsupported message_type"
}
```

## 7. 当前联调相关 HTTP 接口

- `GET /api/sessions`
- `GET /api/sessions/:id`
- `GET /api/sessions/:id/events`
- `POST /api/sessions/:id/commands/prompt`
- `GET /api/sessions/:id/commands`

这些接口应被视为：

- 当前兼容接口
- 迁移中的 legacy API

## 8. 新协议迁移方向

后续 `bridge -> server` 应围绕下面这些对象重构：

- `device_hello`
- `project_snapshot`
- `thread_snapshot`
- `message`
- `thread_state`
- `debug_event`

同时，`server` 的角色应从“中心化内容聚合器”收敛为：

- 设备发现
- 路由协商
- 鉴权
- relay fallback

## 9. 产品层与调试层分离

当前兼容接口中的下列数据，只能属于调试层：

- heartbeat
- raw command line
- observing 文本
- 无业务语义的 payload

产品层应该看到的是：

- project
- thread
- message history
- thread summary
- thread state

## 10. 结论

如果后续实现仍然只是在这个兼容接口上叠加更多 event 展示，那么方向就是错的。

正确方向是：

- 保留这套兼容接口用于迁移
- 但新的产品设计、移动端视图和 server 角色，都转向 `project/thread/message + P2P`
