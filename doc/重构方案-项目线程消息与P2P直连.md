# codeScope 重构方案：项目 / 线程 / 消息 与 P2P 直连
> 更新时间：2026-03-19
> 状态：目标设计，非当前已实现能力

---

## 1. 重构目标

把 codeScope 从当前的：

```text
session/event 监视器
```

重构为：

```text
项目 -> 线程 -> 消息历史 的移动端 AI 对话观察器
```

并且让：

```text
server = P2P 发现与路由服务
bridge <-> mobile = 默认内容直连
```

---

## 2. 为什么必须重构

当前系统的主要问题不是 UI 不够漂亮，而是产品模型错位：

- 首页显示的是 session 或 workspace 名，不是项目
- 点进去看到的是 heartbeat / command line / observing 提示
- 用户看不到真正的输入和输出对话历史
- server 仍被设计成中心聚合器，而不是隐私优先的发现路由层

这会直接导致产品偏离核心需求：

**用户离开电脑后，想在手机上看 AI 在项目里聊了什么、做到了哪一步、接下来等不等自己。**

---

## 3. 新架构总览

```text
独立运行的 Codex CLI / Claude Code
                |
                | 旁路识别项目 / 线程 / 消息
                v
bridge
                |
                | 向 server 注册设备、项目、线程摘要
                v
server
         | discovery / auth / route / relay fallback
         |
mobile ---P2P direct--- bridge
         |
         └-- fallback via server relay
```

---

## 4. 统一领域模型

## 4.1 Device

表示一台运行 bridge 的电脑。

## 4.2 Project

表示某个 workspace 对应的工程项目。

关键要求：

- 首页先看项目，不先看 session
- 项目名不能简单直接等于文件夹名，必须允许 bridge 识别更友好的展示名

## 4.3 Thread

表示某项目里的一个具体 agent 对话：

- Codex GUI 语义：线程
- Claude Code 语义：对话
- 底层统一：`thread`

## 4.4 Message

表示线程内的一条用户输入或 agent 输出。

## 4.5 ThreadState

面向用户的状态，不是底层进程状态：

- `running`
- `waiting_prompt`
- `waiting_review`
- `completed`
- `blocked`

---

## 5. 三端重构方向

## 5.1 bridge 重构

### 当前状态

- 已实现：发现进程、推 session/event、文件变化、heartbeat
- 未实现：线程与消息历史

### 重构目标

bridge 需要增加：

1. `project correlator`
2. `thread correlator`
3. `message capture adapters`
4. `thread state derivation`
5. `p2p transport endpoint`

### 设计原则

- 不托管启动
- 不破坏用户工作流
- 产品层输出优先于调试层输出

## 5.2 server 重构

### 当前状态

- 已实现：session/event 聚合、REST、WS
- 当前身份：中心化 aggregator

### 重构目标

server 改成：

- device registry
- project/thread index
- p2p route broker
- auth/token issuer
- relay fallback

### 设计原则

- 默认不长期持有消息正文
- 默认优先让 bridge/mobile 直接通信
- 只有在直连失败时才中继

## 5.3 mobile 重构

### 当前状态

- 已实现：session/event 浏览、prompt/file 基础页面
- 问题：仍偏 event viewer

### 重构目标

移动端主导航：

1. 设备列表
2. 项目列表
3. 项目详情
4. 线程列表
5. 线程消息历史
6. 项目文件树
7. 新建线程入口

### 设计原则

- 默认显示用户最关心的信息
- debug 事件独立收纳
- 心跳不进 release 主界面

---

## 6. 新的数据面与控制面

## 6.1 数据面

用于用户真正关心的内容：

- project snapshots
- thread snapshots
- messages
- file tree / file preview

## 6.2 控制面

用于交互能力：

- create thread
- send prompt
- interrupt thread
- request refresh

## 6.3 调试面

仅用于开发诊断：

- heartbeat
- raw command
- process observation
- capture errors

---

## 7. P2P 隐私模型

## 7.1 默认模式

`mobile <-> bridge` 直接传输消息历史和文件内容。

## 7.2 server 只做什么

- 登录/设备发现
- 颁发短期 route token
- 交换连接候选
- 必要时做 relay

## 7.3 server 默认不做什么

- 默认存储全部消息正文
- 默认转发所有对话数据
- 默认成为聊天中心

---

## 8. 建议实施阶段

## Phase 1：模型重构

目标：

- 在 `server` 和 `mobile` 中引入 `project/thread/message` 模型
- 旧 `session/event` 继续存在但被标记为 legacy

## Phase 2：bridge 识别线程与状态

目标：

- 从当前旁路采集器升级到“项目/线程/状态”采集器

## Phase 3：消息历史采集

目标：

- 尽可能采到真实的 user/assistant 消息内容

## Phase 4：P2P 发现与直连

目标：

- `server` 切换为 discovery/router
- `bridge/mobile` 建立内容直连

## Phase 5：项目级新建线程与文件体验完善

目标：

- 新建线程
- 项目级文件树
- 更贴近 Codex GUI / Claude 对话体验

---

## 9. 当前文档需要如何理解

### 现状文档

- 描述当前代码真实行为

### 目标设计文档

- 描述未来产品语义

### 实施计划文档

- 描述分阶段执行路径

本文件属于：

**目标设计文档**

---

## 10. 本轮重构后的统一判断标准

今后任何改动，都应先回答：

1. 用户在首页看到的是不是项目，而不是 session
2. 用户进入项目后看到的是不是线程，而不是 event 类型
3. 用户能不能看到真实输入和输出历史
4. 用户能不能知道线程是在运行、结束、等 prompt、还是等 review
5. 内容是否优先 P2P 传输

如果五个问题里有三个以上回答是否定的，就说明改动仍然偏离目标。

---

## 11. 一句话结论

codeScope 下一阶段不是继续补 event UI，而是：

**整体重构成面向项目、线程、消息历史的移动端 AI 对话观察系统，并把 server 降级为 P2P 发现与路由层。**
