# Bridge TODO 清单
> 更新日期：2026-03-19

本文档基于新的目标设计整理：

- `bridge` 是独立运行的 side-channel collector
- 不负责托管或替代 `Codex CLI` / `Claude Code`
- 用户原有 CLI 工作流不变
- `bridge` 的职责是尽量重建 `project / thread / message / thread_state`

## P0

### 项目与线程识别

- 明确如何从 `workspace_root` 和运行中的 CLI 会话识别 `project`
- 明确如何把同一项目中的多条 CLI 会话映射成 `thread`
- 支持 `bridge` 晚于 CLI 启动时的识别
- 支持同机多个项目、多个 thread 并存

### 兼容层保持可用

- 保证 legacy `session/event` 继续稳定上报
- 保证 `bridge` 重连后兼容层仍可恢复
- 明确 ACK、重试与失败日志

### 状态推导

- 推导 `thread_state`
- 区分：
  - `running`
  - `waiting_prompt`
  - `waiting_review`
  - `completed`
  - `blocked`

## P1

### message capture adapters

- 设计 `capture adapter`，把：
  - process discovery
  - terminal attach
  - message extraction
  - state derivation
  拆开
- 为 `Codex CLI` / `Claude Code` 分别制定最小适配策略
- 增加可替换的 mock source，便于脱离真实 CLI 联调

### 文件变化采集

- workspace watcher 继续保留
- 文件变化要能和 `project` / `thread` 对齐
- 处理 burst write 去重和聚合
- 支持可配置 ignore 规则

### 运行可观测性

- 输出明确的发现日志
- 输出明确的 project/thread 归属日志
- 输出明确的采集失败和解析失败日志

## P2

### Prompt 链路重设计

- 明确 `send_prompt` 在“不改变独立 CLI 启动方式”前提下如何落地
- 评估方案：
  - 仅记录并提示用户手动处理
  - 写入共享输入通道
  - 使用工具自身支持的扩展注入机制
- 在没有可靠注入方案前，不假设 bridge 能直接驱动 CLI stdin

### P2P 能力

- 为 `bridge <-> mobile` 直连预留 endpoint
- 支持 route offer / answer
- 设计 relay fallback

### 安全

- `bridge -> server` 增加鉴权
- 限制 workspace 访问边界
- 对日志和错误事件做敏感信息脱敏

## P3

- Windows 打包脚本
- 配置模板
- 后台服务模式
- 更完整的联调与回归测试方案

## 当前最重要的设计约束

- `bridge` 是旁路观察者，不是 launcher
- 不改变 `Codex CLI` / `Claude Code` 的启动方式
- `server` 只依赖标准协议，不依赖 `bridge` 内部采集实现
- 产品主数据不是 heartbeat 和 raw command

## 下一步优先级建议

1. 先完成 `project/thread/message` 的 bridge 内部建模
2. 再补 thread_state 推导
3. 再推进 P2P endpoint 与 route 能力
4. 继续保留 legacy `session/event` 兼容层直到 mobile 完成迁移
