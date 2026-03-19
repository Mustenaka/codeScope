# codeScope

[中文说明 / Chinese README](./README_ZH.md)

codeScope is a cross-device system for checking what your coding agents are doing while you are away from your computer, then continuing the same work from your phone.

The intended product model is:

```text
device -> project -> thread -> message -> thread_state -> file browser
```

Rather than exposing a raw `session/event` debugger as the primary experience, codeScope is being shaped into a mobile-first view of:

- which projects are active on a computer
- which agent threads belong to each project
- what the user sent and what the agent replied
- whether a thread is still running, waiting, completed, blocked, stale, or offline
- which project files can be inspected from mobile

## Repository Layout

- `bridge/`: local side-channel collector that discovers running agent processes and emits compatibility events plus partial high-level semantics
- `server/`: current REST/WebSocket backend and derived `project/thread/message` read model
- `mobile/`: Flutter client for project, thread, message, file browser, and prompt flows
- `doc/`: core product, protocol, roadmap, and maintenance documents
- `docs/`: auxiliary implementation notes generated during development

## Current Status

### Implemented

- `bridge` runs in side-channel `discovery` mode and does not replace the user's normal Codex/Claude workflow.
- `bridge` can discover local `codex` / `claude` processes, watch workspace file changes, and forward compatibility-layer events upstream.
- `bridge` also reads recent Codex local session files and can emit partial real `user` / `assistant` content.
- `server` derives `project`, `thread`, and `message` read APIs from the current ingestion model while keeping legacy `sessions` endpoints available.
- `mobile` already uses a `project -> thread -> message` primary navigation flow backed by REST and WebSocket data.

### Partially Implemented

- `thread` identity still has legacy `session` ancestry in several places.
- `thread_state` still relies on heuristics in parts of the pipeline.
- real message capture is stronger for Codex local sessions than for other agents
- realtime list refresh still mixes incremental updates with snapshot reloads

### Not Yet Implemented

- complete real input/output capture for all supported agents
- project-level create-thread flow
- formal `bridge <-> mobile` P2P direct transport with relay fallback
- fully authoritative thread lifecycle state machine

## Architecture Direction

codeScope follows these constraints:

- `bridge` stays a side-channel collector, not a managed launcher
- `server` should default toward discovery, routing, and optional relay fallback, not a permanent content center
- message bodies, thread history, and file content should eventually prefer `bridge <-> mobile` direct transfer
- heartbeat, raw command line, bridge observing hints, and low-level payloads belong to debug views, not the release main view

## Quick Start

### Run the server

```powershell
cd server
go run ./cmd/server
```

### Run the bridge

```powershell
cd bridge
go run ./cmd/bridge
```

### Run the compatibility test loop without mobile

```powershell
cd bridge
$env:CODESCOPE_BRIDGE_SERVER_URL="ws://localhost:8080/ws/bridge"
$env:CODESCOPE_BRIDGE_SESSION_ID="session-demo"
$env:CODESCOPE_BRIDGE_SOURCE_MODE="jsonl"
go run ./cmd/fake-source -workspace-root "D:/Work/Code/Cross/codeScope" | go run ./cmd/bridge
```

### Run the debug subscriber

```powershell
cd server
go run ./cmd/debug-subscriber -server-url ws://localhost:8080/ws/mobile -session-id session-demo
```

### Check health and APIs

```powershell
curl http://localhost:8080/api/health
curl http://localhost:8080/api/sessions
curl http://localhost:8080/api/projects
```

## Verification

Run verification inside each module:

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

## Documentation

Start with these documents:

1. [`doc/全局修改指南-AI-Agent协作基线.md`](./doc/%E5%85%A8%E5%B1%80%E4%BF%AE%E6%94%B9%E6%8C%87%E5%8D%97-AI-Agent%E5%8D%8F%E4%BD%9C%E5%9F%BA%E7%BA%BF.md)
2. [`doc/重构方案-项目线程消息与P2P直连.md`](./doc/%E9%87%8D%E6%9E%84%E6%96%B9%E6%A1%88-%E9%A1%B9%E7%9B%AE%E7%BA%BF%E7%A8%8B%E6%B6%88%E6%81%AF%E4%B8%8EP2P%E7%9B%B4%E8%BF%9E.md)
3. [`doc/bridge/Bridge旁路采集需求重设计.md`](./doc/bridge/Bridge%E6%97%81%E8%B7%AF%E9%87%87%E9%9B%86%E9%9C%80%E6%B1%82%E9%87%8D%E8%AE%BE%E8%AE%A1.md)
4. [`doc/bridge-server-完整协议草案.md`](./doc/bridge-server-%E5%AE%8C%E6%95%B4%E5%8D%8F%E8%AE%AE%E8%8D%89%E6%A1%88.md)
5. [`doc/正式版本落地路线图.md`](./doc/%E6%AD%A3%E5%BC%8F%E7%89%88%E6%9C%AC%E8%90%BD%E5%9C%B0%E8%B7%AF%E7%BA%BF%E5%9B%BE.md)
6. [`doc/执行提示词-正式版落地.md`](./doc/%E6%89%A7%E8%A1%8C%E6%8F%90%E7%A4%BA%E8%AF%8D-%E6%AD%A3%E5%BC%8F%E7%89%88%E8%90%BD%E5%9C%B0.md)

## Next Steps

1. Complete real message capture and reduce reliance on session-derived thread heuristics.
2. Turn `thread_state` into a proper lifecycle state machine.
3. Finish the mobile prompt continuation loop.
4. Implement `bridge <-> mobile` P2P direct transport with relay fallback.
5. Promote project-level file browser and create-thread flows into the primary experience.
