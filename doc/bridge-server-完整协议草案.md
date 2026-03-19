# bridge-server 协议草案（P2P 发现与路由版）
> 更新时间：2026-03-19

## 1. 目标

本协议从现在开始不再仅服务于 `session/event` 聚合，而是服务于：

- `project`
- `thread`
- `message`
- `thread_state`
- `file_browser`
- `p2p discovery / route / relay fallback`

## 2. 架构定位

### 2.1 server 的新角色

server 默认角色：

- 设备发现服务
- 会话路由服务
- P2P 协商服务
- relay fallback 服务

server 默认**不是**：

- 全量消息中心
- 默认内容存储中心

### 2.2 内容传输优先级

1. `bridge <-> mobile` P2P direct
2. `server relay fallback`
3. `server persisted snapshot` 仅在用户显式允许时

## 3. 顶层协议对象

### 3.1 device

标识一个 bridge 所在电脑。

建议字段：

- `device_id`
- `device_name`
- `bridge_version`
- `agent_kinds`
- `network_candidates`

### 3.2 project

建议字段：

- `project_id`
- `project_name`
- `workspace_root`
- `thread_count`
- `running_thread_count`
- `last_activity_at`

### 3.3 thread

建议字段：

- `thread_id`
- `project_id`
- `agent_kind`
- `title`
- `status`
- `summary`
- `started_at`
- `last_activity_at`
- `ended_at`

### 3.4 message

建议字段：

- `message_id`
- `thread_id`
- `role`
- `content`
- `created_at`
- `sequence`
- `metadata`

## 4. bridge -> server 消息

### 4.1 device_hello

bridge 建立连接后先上报设备信息与可用网络候选。

### 4.2 project_snapshot

bridge 上报当前设备上有哪些项目。

### 4.3 thread_snapshot

bridge 上报某项目下有哪些线程，以及线程状态摘要。

### 4.4 message

bridge 上报线程里的消息。

注意：

- 这类消息未来优先走 P2P
- 通过 server 的路径默认只做发现期 / fallback 期支持

### 4.5 debug_event

保留 heartbeat / error / process_observation / raw_command。

这些消息默认不用于 release 版主界面。

## 5. server -> mobile 消息

### 5.1 device_list

告诉 mobile 当前能看到哪些 bridge 设备。

### 5.2 project_list

告诉 mobile 选中设备上有哪些项目。

### 5.3 thread_list

告诉 mobile 项目下有哪些线程。

### 5.4 route_offer

告诉 mobile 当前应该尝试的连接方式：

- `p2p_direct`
- `relay_fallback`

### 5.5 relay_stream

只有在无法直连时启用。

## 6. mobile -> server 消息

### 6.1 device_subscribe

订阅某台 bridge 设备。

### 6.2 project_subscribe

订阅某个项目。

### 6.3 thread_subscribe

订阅某个线程。

### 6.4 route_request

请求与指定 bridge 建立 P2P 连接。

## 7. P2P 协商消息

server 需要支持的最小协商内容：

- `peer_id`
- `session_token`
- `offer`
- `answer`
- `ice_candidates` 或等价候选地址
- `relay_required`

具体底层可以后续选型：

- WebRTC data channel
- 自定义 TCP/QUIC 打洞
- 局域网直连发现

本草案先不绑定具体技术，但必须保留这些协商位。

## 8. REST API 方向重构

当前 REST 以 `sessions` 为中心，后续应重构为：

- `GET /api/devices`
- `GET /api/devices/:id/projects`
- `GET /api/projects/:id`
- `GET /api/projects/:id/threads`
- `GET /api/threads/:id`
- `GET /api/threads/:id/messages`
- `POST /api/projects/:id/threads`
- `POST /api/threads/:id/commands/prompt`
- `GET /api/projects/:id/files/tree`
- `GET /api/projects/:id/files/content`
- `POST /api/p2p/route`

兼容策略：

- 旧 `sessions` API 可以短期保留
- 但必须在文档中标记为 legacy

## 9. 当前实现与目标实现的差距

### 已实现

- bridge/server websocket 基础链路
- mobile 订阅 server websocket
- file browser 基础 API
- prompt command task 基础 API

### 半实现

- session 作为 thread 的早期替代物
- command_result 作为有限交互结果

### 未实现

- device/project/thread/message 完整模型
- P2P 发现与协商
- P2P 内容直连
- 新建线程接口
- 真实消息历史采集

## 10. 协议演进规则

1. 新产品主模型优先围绕 `project/thread/message`
2. `heartbeat` / `raw_command` / `process_observation` 归入 debug 类
3. `server` 不能默认要求持有全部消息正文
4. 所有与消息正文相关的设计都要优先考虑 P2P

## 11. 一句话结论

这份协议从现在开始服务的不是“中心化 session/event 监控”，而是：

**以 P2P 为默认隐私前提的项目、线程、消息、文件与移动端观察协议。**
