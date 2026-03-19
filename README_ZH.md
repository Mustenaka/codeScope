# codeScope

[English README](./README.md)

codeScope 是一个面向 AI coding agent 的跨设备协作系统。

它的目标不是把手机做成原始 `session/event` 调试面板，而是让你离开电脑后，仍然能在手机上直接看到：

- 哪些工作区正在活跃
- 每个工作区里有哪些对话线程
- 用户发了什么，agent 回复了什么
- 线程当前是运行中、等待下一次 prompt、等待 review、已完成、阻塞、陈旧还是离线
- 哪些项目文件可以从移动端浏览

当前正式版目标模型是：

```text
device -> project -> thread -> message -> thread_state -> file browser
```

也就是说，release 主链路应该围绕 `workspace/project -> thread -> message` 展开；legacy `session` / `event` 页面仍然保留，但主要用于兼容与调试。

## 仓库结构

- `bridge/`：本地旁路采集器；发现运行中的 agent 进程，上报兼容事件，并补充部分语义化采集
- `server/`：REST / WebSocket 后端；从当前 ingestion 流程派生 `project/thread/message` 读模型
- `mobile/`：Flutter 客户端；负责工作区、线程、消息、文件浏览和 prompt 交互
- `doc/`：架构、协议、路线图、维护文档
- `docs/`：开发过程中生成的辅助实现说明

## 当前状态

### 已实现

- `mobile` 主导航已经是 `workspace -> thread -> message`
- `server` 已提供 release 主链路 API：
  - `projects`
  - `project threads`
  - `thread messages`
  - `thread prompt continuation`
  - `project file browser`
  - `project create thread`
- `bridge` 已切到 side-channel discovery 模式，不要求用户改变原有 Codex / Claude 启动方式
- `bridge` 可以发现本机 `codex` / `claude` 进程，并采集兼容层事件
- `bridge` 已有 Codex 与 Claude 的语义化 transcript capture adapter
- `thread_state` 已经超出简单 session 状态，当前支持：
  - `running`
  - `waiting_prompt`
  - `waiting_review`
  - `completed`
  - `blocked`
  - `offline`
  - `stale`
- 当前 release UX 已能展示：
  - 工作区创建时间
  - 工作区最近活跃时间
  - `Codex` / `Claude` 来源标识
  - 长 prompt / 长 summary / 长 message 的抽屉式查看
  - “没有可写 bridge session” 这类 prompt 失败的用户化提示

### 半实现

- `thread` 对外已经是正式主链路的一部分，但执行层底下仍会回落到可写 backing `session`
- 真实消息采集比早期强很多，但还没有覆盖所有 agent、所有 turn 形态、所有消息类型
- `thread_state` 现在已经更像正式状态机，但仍然是 server 侧 read-model 推导，不是独立持久化 runtime state store
- 实时刷新仍然混合使用流式更新和 REST 快照重拉
- 文件内容和消息正文目前仍主要由 `server` 承载，不是未来目标里的 `bridge <-> mobile` 直传优先

### 未实现

- `bridge <-> mobile` 正式 P2P direct 传输
- relay fallback / reconnect / replay
- 所有支持 agent 的完整真实输入 / 输出采集闭环
- 执行层彻底脱离 backing `session`

## 项目约束

这些约束是刻意保留的，不应回退：

- `bridge` 必须保持旁路采集器定位，而不是托管启动器
- `server` 默认应是 discovery / routing / optional relay，而不是永久内容中心
- heartbeat、原始命令行、bridge observing 提示、底层 payload 只属于 debug 视图，不属于 release 主视图
- 消息正文、线程历史、文件内容未来应优先走 `bridge <-> mobile` 直接传输

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

### 运行 mobile

```powershell
cd mobile
flutter run
```

### 截图

![e4d9ef818f9645193e64c6330c33ff80](//pic//e4d9ef818f9645193e64c6330c33ff80.jpg)

![1ee45df453efa89354df6a74d28d8fd9](//pic//1ee45df453efa89354df6a74d28d8fd9.jpg)

![6f47c8831d12c8f0efe35a95f09d38f2](//pic//6f47c8831d12c8f0efe35a95f09d38f2.jpg)

![53bfb44f1b882d10a09b762dae63923f](//pic//53bfb44f1b882d10a09b762dae63923f.jpg)

![a7ad9c2fb6e3ea7be198cfc7c3ca097e](//pic//a7ad9c2fb6e3ea7be198cfc7c3ca097e.jpg)

![d3ebc99456a9bc9f27c51128eec47b03](//pic//d3ebc99456a9bc9f27c51128eec47b03.jpg)

### 常用接口检查

```powershell
curl http://localhost:8080/api/health
curl http://localhost:8080/api/projects
curl http://localhost:8080/api/projects/<project-id>/threads
curl http://localhost:8080/api/threads/<thread-id>/messages
```

## 验证

分别在各模块目录执行：

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

## 核心文档入口

建议从这些文档开始：

1. [`doc/全局修改指南-AI-Agent协作基线.md`](./doc/%E5%85%A8%E5%B1%80%E4%BF%AE%E6%94%B9%E6%8C%87%E5%8D%97-AI-Agent%E5%8D%8F%E4%BD%9C%E5%9F%BA%E7%BA%BF.md)
2. [`doc/重构方案-项目线程消息与P2P直连.md`](./doc/%E9%87%8D%E6%9E%84%E6%96%B9%E6%A1%88-%E9%A1%B9%E7%9B%AE%E7%BA%BF%E7%A8%8B%E6%B6%88%E6%81%AF%E4%B8%8EP2P%E7%9B%B4%E8%BF%9E.md)
3. [`doc/bridge/Bridge旁路采集需求重设计.md`](./doc/bridge/Bridge%E6%97%81%E8%B7%AF%E9%87%87%E9%9B%86%E9%9C%80%E6%B1%82%E9%87%8D%E8%AE%BE%E8%AE%A1.md)
4. [`doc/bridge-server-完整协议草案.md`](./doc/bridge-server-%E5%AE%8C%E6%95%B4%E5%8D%8F%E8%AE%AE%E8%8D%89%E6%A1%88.md)
5. [`doc/正式版本落地路线图.md`](./doc/%E6%AD%A3%E5%BC%8F%E7%89%88%E6%9C%AC%E8%90%BD%E5%9C%B0%E8%B7%AF%E7%BA%BF%E5%9B%BE.md)
6. [`doc/执行提示词-正式版落地.md`](./doc/%E6%89%A7%E8%A1%8C%E6%8F%90%E7%A4%BA%E8%AF%8D-%E6%AD%A3%E5%BC%8F%E7%89%88%E8%90%BD%E5%9C%B0.md)

## 下一步重点

1. 继续补齐更多 agent transcript 形态下的真实消息采集
2. 继续减少执行层对 backing `session` 的依赖
3. 把消息正文和文件内容逐步迁到 `bridge <-> mobile` 直传优先
4. 最后再落地 P2P direct 与 relay fallback
