# codeScope 文档索引

当前 `doc` 目录已经以 `2026-03-19` 的新基线为准。

## 优先阅读顺序

1. [全局修改指南-AI-Agent协作基线.md](/D:/Work/Code/Cross/codeScope/doc/全局修改指南-AI-Agent协作基线.md)
2. [重构方案-项目线程消息与P2P直连.md](/D:/Work/Code/Cross/codeScope/doc/重构方案-项目线程消息与P2P直连.md)
3. [bridge/Bridge旁路采集需求重设计.md](/D:/Work/Code/Cross/codeScope/doc/bridge/Bridge旁路采集需求重设计.md)
4. [bridge-server-完整协议草案.md](/D:/Work/Code/Cross/codeScope/doc/bridge-server-完整协议草案.md)
5. [正式版本落地路线图.md](/D:/Work/Code/Cross/codeScope/doc/正式版本落地路线图.md)
6. [执行提示词-正式版落地.md](/D:/Work/Code/Cross/codeScope/doc/执行提示词-正式版落地.md)

## 各文档职责

- `全局修改指南-AI-Agent协作基线.md`
  统一产品基线和跨端约束。明确产品主模型是 `device -> project -> thread -> message -> thread_state`，不是 `session/event` 监视器。
- `重构方案-项目线程消息与P2P直连.md`
  跨 `bridge / server / mobile` 的重构方向文档，解释为什么不能继续沿着 event viewer 修补。
- `bridge/Bridge旁路采集需求重设计.md`
  约束 `bridge` 继续保持旁路采集，不回到托管启动；同时把目标从“上报低层事件”改成“尽量重建项目、线程、消息和状态”。
- `bridge-server-完整协议草案.md`
  定义未来协议应服务于 `project/thread/message` 和 `p2p discovery / route / relay fallback`。
- `正式版本落地路线图.md`
  面向正式版本的跨端实施路线图，明确当前状态、阶段目标、受影响模块、风险和推荐顺序。
- `执行提示词-正式版落地.md`
  后续继续推进实现时使用的统一提示词集，覆盖 bridge、server、mobile、协议、文档治理和测试。
- `后续修改统一提示词-项目线程消息与P2P版.md`
  兼容保留的总提示词入口；后续优先使用 `执行提示词-正式版落地.md` 中的更细分提示词。

## 当前保留的兼容层文档

下面这些文档仍可保留，但只能按“兼容层 / 联调 / 历史迁移文档”理解：

- `脱离mobile联调测试闭环.md`
- `bridge-server-接口对接与测试.md`
- `bridge/Bridge-Server接口说明.md`
- `bridge/TODO清单.md`

这些文档如果提到以下内容，只能按“历史实现/兼容层”理解：

- `session`
- `event`
- `heartbeat`
- `terminal_output`
- `Monitor Server`
- `Session 列表`
- 以 `server` 为默认内容中心

## 已删除的重复文档

以下文档已被更高层级的新基线和正式落地路线图完全覆盖，因此删除：

- `技术方案.md`
- `实施规划.md`
- `项目设计文档.md`
- `bridge/Server实现提示词.md`
- `prompt/bridge`
- `prompt/bridge-server`
- `prompt/mobile`
- `prompt/server`
- `prompt/test`

## 当前整理结论

- 产品目标已经从“远程查看底层 bridge/server 事件”切换成“远程查看 PC 上 coding agent 的项目、线程、消息历史、线程状态和项目文件”。
- `mobile` release 视图默认不应该显示 heartbeat、原始命令行和 bridge observing 文本。
- `server` 的默认身份是 P2P 发现与路由服务，不是默认内容中心。
- 消息正文、线程历史、文件内容优先考虑 `bridge <-> mobile` 直连；`server relay` 只做 fallback。
