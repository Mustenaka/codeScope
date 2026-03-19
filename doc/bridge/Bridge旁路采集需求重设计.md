# Bridge 旁路采集需求重设计
> 更新时间：2026-03-19

## 1. 新前提

bridge 的核心原则不变：

- 用户仍然按原方式独立运行 `Codex CLI` / `Claude Code`
- bridge 仍然不能把产品改回“托管启动器”
- bridge 仍然是独立常驻的旁路采集器

但 bridge 的目标从“上报底层事件”升级为：

**尽可能旁路重建项目、线程、消息历史与用户态状态。**

## 2. bridge 真正要解决的问题

bridge 不只是告诉系统“本机有个 claude.exe 在跑”，而是尽可能回答：

1. 这属于哪个项目
2. 这是项目里的哪个线程/对话
3. 用户刚刚发了什么
4. agent 刚刚输出了什么
5. 这个线程现在是运行中、等待 prompt、等待确认、还是结束

## 3. 不再把哪些东西当作产品成功标准

以下只能算调试信息，不能算产品闭环：

- heartbeat
- 原始进程命令行
- `[bridge] observing ...`
- 单纯的 process discovery

这些信息可以保留，但必须降级到 debug 诊断层。

## 4. bridge 的双层输出模型

## 4.1 产品层输出

bridge 未来应优先输出：

- `project_snapshot`
- `thread_snapshot`
- `message`
- `thread_state`
- `file_change`

## 4.2 调试层输出

bridge 继续保留：

- `heartbeat`
- `error`
- `process_observation`
- `raw_command`

## 5. 采集优先级

### 第一优先级

项目 / 线程识别：

- 哪个 workspace 对应一个项目
- 一个项目下有几个活跃线程
- 每个线程和哪个 agent_kind 对应

### 第二优先级

消息历史识别：

- 用户输入
- assistant 输出
- 系统提示

### 第三优先级

用户态状态识别：

- 运行中
- 等待用户 prompt
- 等待用户确认 / review
- 已结束
- 阻塞 / 错误

### 第四优先级

调试辅助：

- heartbeat
- 文件变化
- process lifecycle

## 6. 推荐 bridge 内部结构

```text
Discovery Layer
  -> Project/Thread Correlator
  -> Message Capture Adapters
  -> State Derivation Layer
  -> Debug Event Layer
  -> P2P Transport / Relay Fallback Layer
```

## 7. 与 server 的关系

bridge 与 server 的关系要重新定义：

- server 默认不是 bridge 的内容中心
- server 默认只负责发现、协商、路由
- bridge 最终应优先把真实消息历史发给 mobile 的 P2P 连接
- server 只在不可直连时做 fallback relay

## 8. 当前实现状态判断

### 已实现

- 旁路发现 `codex` / `claude`
- workspace_root 关联
- side-channel 默认路径
- 文件变化
- heartbeat

### 半实现

- session 稳定映射
- prompt task / command_result

### 未实现

- thread 模型
- message 历史
- 用户输入/输出真实捕获
- P2P 内容直连

## 9. 后续 bridge 修改硬规则

1. 不得把 bridge 改回托管启动模式
2. 不得再把 heartbeat/原始命令行视为主产品数据
3. 后续采集器优先服务 `project/thread/message` 模型
4. 任何新增协议字段都要优先考虑 P2P 传输，而不是默认交给中心 server 长期持有

## 10. 一句话结论

bridge 的正确方向不是“把进程事件发得更全”，而是：

**在不改变用户工作流的前提下，尽可能旁路重建项目、线程、消息和状态，并优先把这些内容通过 P2P 提供给 mobile。**
