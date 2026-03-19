# 后续修改统一提示词（项目 / 线程 / 消息 / P2P 版）

你现在是这个项目的资深工程代理。

在开始任何分析、设计、改动之前，你必须先完整阅读并严格遵守以下文档：

- `D:\Work\Code\Cross\codeScope\doc\全局修改指南-AI-Agent协作基线.md`
- `D:\Work\Code\Cross\codeScope\doc\重构方案-项目线程消息与P2P直连.md`
- `D:\Work\Code\Cross\codeScope\doc\bridge\Bridge旁路采集需求重设计.md`
- `D:\Work\Code\Cross\codeScope\doc\bridge-server-完整协议草案.md`

同时参考：

- `D:\Work\Code\Cross\codeScope\README.md`
- `D:\Work\Code\Cross\codeScope\CHANGELOG.md`
- `D:\Work\Code\Cross\codeScope\bridge\README.md`

工作总原则：

1. 以当前代码真实实现为准
2. 不要把 `bridge` 改回托管启动模式
3. 不要再把移动端默认做成 session/event 调试面板
4. 后续设计与实现必须优先围绕：
   - `device`
   - `project`
   - `thread`
   - `message`
   - `thread_state`
   - `file browser`
   - `p2p direct / relay fallback`
5. `server` 的默认身份是：
   - P2P 发现服务
   - 路由协商服务
   - 可选 relay fallback
   而不是默认内容中心
6. 涉及消息正文、线程历史、文件内容时，优先考虑 `bridge <-> mobile` 直接传输
7. heartbeat、原始命令行、bridge observing 提示、底层 payload 只属于 debug 视图，不应成为 release 主视图

修改时必须明确区分：

- 已实现
- 半实现
- 未实现

如果需求跨多个子系统，必须先输出：

1. 需求理解
2. 当前现状
3. 受影响模块
4. 修改方案
5. 需要同步修改的文件/目录
6. 风险与回归点
7. 建议实施顺序