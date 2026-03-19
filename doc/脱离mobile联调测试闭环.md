# codeScope 脱离 Mobile 的联调测试闭环
> 更新日期：2026-03-19
> 性质：兼容层联调与迁移准备文档

## 1. 文档目的

在 `mobile` 尚未完成或尚未切到新模型时，为 `bridge` 和 `server` 提供一套可独立执行的联调测试方案。

这份文档要验证的是：

```text
测试输入源 -> bridge -> server -> debug subscriber / REST
```

它用于验证当前兼容链路是否可用，并为后续 `project/thread/message + P2P` 迁移预留测试入口。

## 2. 这份文档不代表什么

它不代表：

- `session/event` 就是最终产品模型
- heartbeat 和 raw command 就是 release 成功标准
- server 会长期作为正文内容中心

真正的产品目标仍然是：

- project
- thread
- message history
- thread_state
- file browser
- bridge 与 mobile 优先 P2P 直连

## 3. 当前测试范围

### 当前兼容层

本方案当前重点覆盖：

- `bridge` 兼容事件产生
- `bridge -> server` WebSocket 上报
- `server` 对 legacy `session/event` 的 ingest
- `server` 的 REST 查询能力
- `server` 的实时广播能力

### 迁移预留

本方案也要为后续这些对象预留可测试入口：

- `device`
- `project`
- `thread`
- `message`
- `thread_state`
- `p2p route`

## 4. 测试策略

采用“假输入源 + 真 bridge + 真 server + 假订阅端”的方式做联调：

- 用假输入源替代真实 agent 消息来源
- 用真实 `bridge` 接入真实 `server`
- 用轻量调试订阅器替代 `mobile`

## 5. 闭环结构

```text
fake source
   |
   v
bridge
   |
   v
server
  |      \
  |       \
  v        v
REST API   debug subscriber
```

## 6. 测试工具

### 6.1 fake source

建议位置：

```text
bridge/cmd/fake-source/
```

作用：

- 模拟兼容层输入
- 生成标准化测试消息
- 驱动 `bridge` 走真实上报链路

当前兼容层建议支持：

- `terminal_output`
- `ai_output`
- `file_change`
- `heartbeat`
- `error`
- `session_id`

后续迁移建议逐步补：

- `project_snapshot`
- `thread_snapshot`
- `message`
- `thread_state`

### 6.2 debug subscriber

建议位置：

```text
server/cmd/debug-subscriber/
```

作用：

- 模拟移动端订阅
- 在终端中输出接收到的兼容层或迁移层消息

当前至少支持：

- 连接 server WebSocket
- 可按 `session_id` 过滤兼容消息
- 打印类型、时间、摘要内容

后续建议扩展：

- 按 `project_id`
- 按 `thread_id`
- 按 `message role`

## 7. 当前兼容层测试场景

### 7.1 正常连接

验证：

- `bridge` 能连接 `server`
- `server` 能识别连接并建立兼容 session

### 7.2 连续事件上报

验证：

- `bridge` 可持续发送兼容事件
- `server` 可接收并查询

### 7.3 实时广播

验证：

- `debug subscriber` 能实时收到消息

### 7.4 断线重连

验证：

- `bridge` 与 `server` 断开后能恢复

### 7.5 非法消息处理

验证：

- `server` 对坏消息有韧性

### 7.6 多源并发

验证：

- `server` 能区分多个兼容 session 来源

## 8. 当前兼容层最小接口

### REST

- `GET /api/health`
- `GET /api/sessions`
- `GET /api/sessions/:id`
- `GET /api/sessions/:id/events`

### WebSocket

- `GET /ws/bridge`
- `GET /ws/mobile`

说明：

- 这些接口只是当前联调闭环需要
- 不代表最终产品长期只围绕 `sessions/events`

## 9. 推荐执行顺序

1. 启动 `server`
2. 启动 `debug-subscriber`
3. 启动 `bridge`
4. 启动 `fake-source`
5. 观察实时输出
6. 再查 REST

## 10. 验收标准

### 当前兼容层验收

- `server` 健康检查可用
- `bridge` 能接入并上报
- `server` 能存储和查询兼容层 `session/event`
- `debug-subscriber` 能收到广播
- command task / command_result 闭环可验证

### 迁移准备验收

- 测试工具结构没有绑死在 `session/event`
- 后续可以无大改地补 `project/thread/message` 测试输入和订阅过滤

## 11. 结论

这份闭环方案的价值在于：

- 让当前兼容链路持续可验证
- 避免把 mobile 是否完成当成后端联调前提
- 为新的 `project/thread/message + P2P` 迁移保留测试入口

它不是为了证明“用户就该看 Session/Event”，而是为了让迁移过程有稳定的验证抓手。
