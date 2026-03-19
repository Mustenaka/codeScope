# codeScope 全局修改指南（AI Agent 协作基线）
> 更新时间：2026-03-19
> 适用范围：`bridge` / `server` / `mobile` / 文档 / 测试

---

## 1. 这份基线解决什么问题

从 `2026-03-19` 起，codeScope 的产品目标不再被表述为“移动端查看 bridge/server 事件流”，而是明确表述为：

**用户临时离开电脑后，用手机查看自己 PC 上运行的 AI coding agent 的项目、对话线程、输入输出历史、运行状态，以及项目文件结构。**

这意味着：

- `session/event/heartbeat/terminal_output` 只是当前实现的底层技术形态，不是最终产品模型。
- 移动端默认视角必须贴近 `Codex GUI` / `Claude Code` 的对话视角，而不是 PTY/bridge 调试视角。
- 后续任何全局性修改，都必须围绕 `project -> thread -> message` 的产品语义设计。

---

## 2. 新的产品基线

### 2.1 用户真正关心的信息

用户打开移动端时，核心关注点按优先级排序是：

1. 这台电脑上当前有哪些 AI 项目
2. 每个项目下有哪些对话线程
3. 每个线程是运行中、已结束、等待用户下一条 prompt、还是等待用户确认检查
4. 每个线程里用户发了什么、agent 回了什么
5. 这个项目的代码文件结构，以及代码文件的只读预览
6. 如何基于这个项目发起新对话

### 2.2 默认不应成为主视图的信息

以下信息不应成为 release 版主视图的默认内容：

- heartbeat
- bridge 自己生成的 observing 提示
- 原始进程命令行，如 `claude.exe --output-format stream-json ...`
- 没有产品语义的底层事件 payload

这些信息只能作为：

- debug 视图
- 诊断模式
- 开发联调辅助信息

---

## 3. 当前真实状态与新的解释方式

### 3.1 当前真实实现

当前代码真实实现仍然是：

```text
独立运行的 Codex CLI / Claude Code
                |
                | 旁路发现 / 旁路采集
                v
bridge
                |
                | WebSocket
                v
server
                |
                | REST / WebSocket
                v
mobile
```

### 3.2 必须明确承认的事实

当前 `2026-03-19` 的代码还**没有**真正采集到完整的“对话输入/输出历史”，目前采到的主要仍是：

- 进程发现
- 进程命令行
- 文件变化
- heartbeat
- error
- prompt command task / command_result

因此：

- 当前系统**已实现**：发现本机有多少个 agent 会话
- 当前系统**半实现**：按 session 级别做远端观察与文件浏览
- 当前系统**未实现**：真正的项目/线程/消息历史模型

后续所有设计与代码修改，必须清楚区分这三种状态，不能把“未来目标”表述成“当前已具备能力”。

---

## 4. 新的统一领域模型

从现在开始，跨端设计统一围绕以下模型：

### 4.1 Project

表示“这台电脑上的一个工程项目”。

最小字段建议：

- `project_id`
- `project_name`
- `workspace_root`
- `agent_kinds`
- `thread_count`
- `running_thread_count`
- `waiting_prompt_count`
- `waiting_review_count`
- `last_activity_at`

### 4.2 Thread

表示某个项目中的一个具体对话。

对于 `Codex`，UI 语义可以叫“线程”；
对于 `Claude Code`，UI 语义可以叫“对话”；
但底层模型统一叫 `thread`。

最小字段建议：

- `thread_id`
- `project_id`
- `agent_kind`
- `title`
- `status`
- `summary`
- `started_at`
- `last_activity_at`
- `ended_at`
- `source_session_id`

### 4.3 Message

表示线程中的一条用户输入或 agent 输出。

最小字段建议：

- `message_id`
- `thread_id`
- `role`：`user` / `assistant` / `system`
- `content`
- `created_at`
- `sequence`
- `status`
- `metadata`

### 4.4 Debug Event

保留现有 event 模型，但降级为“诊断数据”，不再作为 release 主模型。

---

## 5. 三端职责重定义

## 5.1 bridge

bridge 的产品定位仍然成立：

- 不托管启动 `codex` / `claude`
- 不改变用户工作流
- 作为旁路采集器独立运行

但 bridge 的目标采集物从“事件”升级为“两层数据”：

1. **产品层数据**
   - 项目识别
   - 线程识别
   - 消息识别
   - 用户态状态识别
2. **调试层数据**
   - heartbeat
   - error
   - 原始命令行
   - 观察提示

## 5.2 server

server 从现在开始不再被定义为“中心化会话聚合器”，而是：

**P2P 发现与路由服务**

默认职责：

- 帮助 `bridge` 与 `mobile` 建立发现关系
- 做最小必要的身份交换 / 会话引导 / NAT 辅助路由
- 在无法直连时，才作为可选 relay / fallback

默认不应成为：

- 长期持有全部对话内容的中心化服务器
- 默认中转所有私密消息的中心节点

## 5.3 mobile

mobile 默认产品视图必须是：

```text
设备
  -> 项目列表
      -> 项目详情
          -> 线程列表
              -> 线程消息历史
          -> 项目文件树
          -> 新建线程入口
```

不是：

```text
session 列表
  -> event 流
      -> heartbeat / terminal / command payload
```

---

## 6. P2P 隐私基线

这是从现在开始的新架构原则：

### 6.1 默认原则

- `bridge` 与 `mobile` 应优先直接 P2P 连接
- `server` 只做发现、认证引导、地址交换、协商
- 用户真实对话内容应尽量只在 `bridge <-> mobile` 之间传输

### 6.2 server 的新身份

server 应被设计成：

- 设备发现服务
- 会话路由服务
- 可选中继服务

不是：

- 默认消息中心
- 默认内容存储中心

### 6.3 隐私级别

建议后续协议里明确三种模式：

- `p2p_direct`
- `relay_fallback`
- `local_only`

移动端需要明确看到当前连接模式。

---

## 7. 从现在开始哪些文档算“现状基线”

优先级从高到低：

1. 当前代码真实实现
2. 本文档
3. 新的重构方案文档：
   - `doc/重构方案-项目线程消息与P2P直连.md`
4. 相关模块 README
5. 历史设计草案

如果历史文档和本文冲突，以本文为准。

---

## 8. 后续修改的硬规则

### 8.1 不允许再把移动端默认视图做成事件监视器

除非用户明确要求 debug 面板，否则不能再把以下内容作为主界面默认信息：

- heartbeat
- 原始 command line
- bridge observing 文本
- 低语义 payload

### 8.2 不允许把当前“未实现的消息历史采集”描述成已实现

必须明确区分：

- 已实现
- 半实现
- 未实现

### 8.3 涉及以下主题时必须三端联动考虑

- `project`
- `thread`
- `message`
- `status`
- `summary`
- `new conversation`
- `file browser`
- `p2p / relay`

### 8.4 任何改动都要回答四个问题

1. 用户看到的项目是什么
2. 用户看到的线程是什么
3. 用户看到的输入输出历史从哪里来
4. 内容是否默认走 P2P

---

## 9. 推荐阅读顺序

### 如果要改总体架构

1. `doc/全局修改指南-AI-Agent协作基线.md`
2. `doc/重构方案-项目线程消息与P2P直连.md`
3. `doc/bridge/Bridge旁路采集需求重设计.md`
4. `doc/bridge-server-完整协议草案.md`
5. 对应代码实现

### 如果要改 bridge

1. `doc/bridge/Bridge旁路采集需求重设计.md`
2. `bridge/README.md`
3. `bridge/internal/discovery/`
4. `bridge/internal/capture/`

### 如果要改 server

1. `doc/重构方案-项目线程消息与P2P直连.md`
2. `doc/bridge-server-完整协议草案.md`
3. `server/internal/http/`
4. `server/internal/session/`
5. `server/internal/event/`

### 如果要改 mobile

1. `doc/重构方案-项目线程消息与P2P直连.md`
2. `mobile/lib/modules/session/`
3. `mobile/lib/modules/log/`
4. `mobile/lib/modules/prompt/`
5. `mobile/lib/modules/file/`

---

## 10. 一句话结论

从 `2026-03-19` 开始，codeScope 的正确方向是：

**把系统从 “session/event 监视器” 重构成 “project/thread/message 的移动端 AI 对话观察器”，并让 `server` 默认承担 P2P 发现与路由角色，而不是默认内容中心。**
