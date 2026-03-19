# codeScope

[中文说明 / Chinese README](./README_ZH.md)

codeScope is a cross-device companion for coding agents.

It lets you leave your computer, open your phone, and still see:

- which workspaces are active on a device
- which threads belong to each workspace
- what the user sent and what the agent replied
- whether a thread is running, waiting, blocked, completed, stale, or offline
- which project files can be browsed from mobile

The intended release model is:

```text
device -> project -> thread -> message -> thread_state -> file browser
```

The primary product is not a raw `session/event` debugger. Legacy `session` and `event` views still exist for compatibility and debugging, but the release direction is workspace/thread/message first.

## Repository Layout

- `bridge/`: local side-channel collector; discovers running agent processes and emits compatibility events plus partial semantic capture
- `server/`: REST/WebSocket backend; derives `project/thread/message` read models from the current ingestion pipeline
- `mobile/`: Flutter client; release-facing workspace, thread, message, file browser, and prompt flows
- `doc/`: architecture, roadmap, protocol, and maintenance documents
- `docs/`: auxiliary implementation notes generated during development

## Current Status

### Implemented

- `mobile` primary navigation is already `workspace -> thread -> message`.
- `server` exposes release-facing APIs for:
  - `projects`
  - `project threads`
  - `thread messages`
  - `thread prompt continuation`
  - `project file browser`
  - `project create thread`
- `bridge` runs in side-channel discovery mode and does not require changing the user’s normal Codex or Claude startup workflow.
- `bridge` can discover local `codex` / `claude` processes and collect compatibility-layer events.
- `bridge` has semantic capture adapters for Codex and Claude transcript sources.
- `thread_state` is already promoted beyond simple session status and supports:
  - `running`
  - `waiting_prompt`
  - `waiting_review`
  - `completed`
  - `blocked`
  - `offline`
  - `stale`
- release UX now shows:
  - workspace created time
  - workspace last activity time
  - source agent badges such as `Codex` and `Claude`
  - long-content drawers for long prompt/summary/message text
  - friendlier prompt failure guidance when no writable local bridge session is online

### Partially Implemented

- `thread` execution is release-facing, but it still resolves to a writable backing `session` under the hood.
- real message capture is much better than before, but it is still incomplete across all agents and all message shapes.
- `thread_state` is now a much clearer read-model state machine, but not yet a fully independent persisted runtime state store.
- realtime list refresh still mixes stream updates with REST snapshot reloads.
- file content and message bodies are still primarily served through `server`, not yet through direct `bridge <-> mobile` transport.

### Not Yet Implemented

- formal `bridge <-> mobile` P2P direct transport
- relay fallback, reconnect, and replay
- fully complete real input/output capture for every supported agent and turn shape
- complete removal of the remaining execution-layer dependency on backing `session`

## Product Constraints

These constraints are intentional and should not be reversed:

- `bridge` stays a side-channel collector, not a managed launcher
- `server` should default toward discovery, routing, and optional relay fallback, not a permanent content center
- heartbeat, raw command lines, bridge observing hints, and low-level payloads belong to debug views, not the release main view
- message bodies, thread history, and file content should eventually prefer `bridge <-> mobile` direct transfer

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

### Run the mobile app

```powershell
cd mobile
flutter run
```

### Useful API checks

```powershell
curl http://localhost:8080/api/health
curl http://localhost:8080/api/projects
curl http://localhost:8080/api/projects/<project-id>/threads
curl http://localhost:8080/api/threads/<thread-id>/messages
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

## Core Documentation

Start with these files:

1. [`doc/全局修改指南-AI-Agent协作基线.md`](./doc/%E5%85%A8%E5%B1%80%E4%BF%AE%E6%94%B9%E6%8C%87%E5%8D%97-AI-Agent%E5%8D%8F%E4%BD%9C%E5%9F%BA%E7%BA%BF.md)
2. [`doc/重构方案-项目线程消息与P2P直连.md`](./doc/%E9%87%8D%E6%9E%84%E6%96%B9%E6%A1%88-%E9%A1%B9%E7%9B%AE%E7%BA%BF%E7%A8%8B%E6%B6%88%E6%81%AF%E4%B8%8EP2P%E7%9B%B4%E8%BF%9E.md)
3. [`doc/bridge/Bridge旁路采集需求重设计.md`](./doc/bridge/Bridge%E6%97%81%E8%B7%AF%E9%87%87%E9%9B%86%E9%9C%80%E6%B1%82%E9%87%8D%E8%AE%BE%E8%AE%A1.md)
4. [`doc/bridge-server-完整协议草案.md`](./doc/bridge-server-%E5%AE%8C%E6%95%B4%E5%8D%8F%E8%AE%AE%E8%8D%89%E6%A1%88.md)
5. [`doc/正式版本落地路线图.md`](./doc/%E6%AD%A3%E5%BC%8F%E7%89%88%E6%9C%AC%E8%90%BD%E5%9C%B0%E8%B7%AF%E7%BA%BF%E5%9B%BE.md)
6. [`doc/执行提示词-正式版落地.md`](./doc/%E6%89%A7%E8%A1%8C%E6%8F%90%E7%A4%BA%E8%AF%8D-%E6%AD%A3%E5%BC%8F%E7%89%88%E8%90%BD%E5%9C%B0.md)

## What To Work On Next

1. Complete real message capture across more agent transcript shapes.
2. Continue reducing execution-layer dependence on backing `session`.
3. Move message/file content toward direct `bridge <-> mobile` transport.
4. Implement P2P direct transport with relay fallback.
