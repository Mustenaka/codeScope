## [0.8.4] - 2026-03-19
### Features
- Added release-facing workspace metadata so project cards and detail pages now expose `created_at` and `last_activity_at` instead of only showing name and path.
- Added agent-source metadata to thread messages so assistant replies can be labeled as `Codex` or `Claude` in the release conversation view.
- Reworked the mobile release UI from `Projects` wording toward clearer `Workspaces` wording and added richer overview cards for workspace/thread context.
- Added long-content handling for release pages: thread summaries, assistant replies, and prompt-task text now collapse into preview blocks with scrollable bottom-sheet detail views.
- Added user-facing server error presentation for missing writable bridge sessions so create-thread and prompt failures now explain how to recover instead of exposing raw transport exceptions.

### Design Rationale
- The formal release flow had already moved to `project -> thread -> message`, but the visible UX still looked like an internal compatibility tool because key context such as time, source agent, and actionable failure guidance was missing.
- Long prompt or summary text should not push core conversation content off screen, so release views now favor preview-first layouts with explicit detail drawers rather than unbounded inline text.
- Users reason about active workspaces, recent activity, and which agent produced a reply more easily than about raw session identifiers or transport exceptions.

### Notes & Caveats
- These changes improve release UX and metadata visibility, but they do not remove the underlying dependency on bridge-online writable sessions for prompt dispatch.
- The app now presents `project has no writable bridge session` as a recoverable local-agent availability issue, but the runtime still requires a bridge-connected executor.
- P2P direct transport, relay fallback, and content-plane migration remain intentionally out of scope for this version.

## [0.8.3] - 2026-03-19
### Features
- Extended the release `thread_state` model with explicit `offline` and `stale` states across `bridge`, `server`, and `mobile`.
- Reworked server-side thread lifecycle derivation into a clearer state-machine flow: latest explicit lifecycle hint first, legacy session status only as fallback, then connectivity/freshness overlays.
- Added bridge-side `waiting_review` lifecycle input for Claude transcript capture when the assistant stops in `pause_turn`.
- Updated mobile thread mapping and summaries so new lifecycle states render as release-oriented thread states instead of falling back to debug/session interpretations.
- Removed default `thread_state=running` tagging from bridge heartbeat and process-observation debug payloads so release lifecycle state is no longer polluted by discovery/debug traffic.

### Design Rationale
- The earlier read model still mixed lifecycle meaning with compatibility-layer heuristics, which made old quiet threads look like `waiting_prompt` and disconnected threads look indistinguishable from active ones.
- `offline` and `stale` are release-facing user states, not transport/debug metadata, so they need to exist in the shared model before any later P2P work.
- Treating bridge lifecycle hints as authoritative inputs and applying connectivity/freshness as a final overlay is more stable than scattering special cases through thread aggregation.

### Notes & Caveats
- The lifecycle model is now more formal, but it is still a derived read-model state machine rather than a fully independent persisted runtime state store.
- `waiting_review` currently has one formal producer path through Claude `pause_turn`; other agents still need equivalent lifecycle hints.
- Non-observed real terminal output can still produce `running` as a lifecycle hint; only debug/heartbeat observation traffic is stripped of release-state semantics.
- P2P direct transport, relay fallback, and content-plane migration remain intentionally out of scope for this version and are still pending.

## [0.8.2] - 2026-03-19
### Features
- Added a dedicated formal-release roadmap document and a consolidated prompt set for future cross-end implementation work.
- Cleaned the documentation set by promoting a smaller core set of baseline documents and removing redundant prompt/plan/design files that repeated the same constraints.
- Updated root and bridge READMEs so they describe the current `project/thread/message` compatibility model instead of the earlier mobile-out-of-scope or event-only framing.

### Design Rationale
- The repository had accumulated multiple overlapping design/plan/prompt documents that said almost the same thing with slightly different emphasis, which increased the risk of future agents drifting back toward a `session/event` observer model.
- A formal release roadmap is more useful than several parallel “design / plan / tech” documents when the main task is coordinated execution across `bridge`, `server`, `mobile`, and documentation.
- Keeping one prompt set aligned with the latest baseline reduces rework and prevents older subsystem-specific prompts from reintroducing outdated assumptions.

### Notes & Caveats
- This version records a documentation-governance cleanup and does not by itself complete the remaining runtime gaps such as full message capture, strict thread lifecycle state, or P2P direct transport.
- Legacy compatibility documents for `sessions/events` testing remain in `doc/` and should still be read as migration/diagnostic material rather than primary product definition.

## [0.8.1] - 2026-03-19
### Features
- Added server-side `project`, `thread`, and `message` read APIs derived from the existing `session/event` ingestion model while keeping legacy session endpoints intact.
- Switched the mobile app primary navigation from a session list into a `project -> thread -> message` flow backed by both mock data and real server REST endpoints.
- Added mobile-side models, mappers, and tests for project lists, thread lists, thread state, and message history.
- Enriched bridge-side event payloads with derived `project_id`, `project_name`, `thread_id`, `source_session_id`, and initial `thread_state` fields without changing the standalone side-channel capture model.
- Added thread-scoped mobile websocket subscription support and made the server thread read model prefer bridge-provided `thread_id` and `thread_state` when available.
- Added project-scoped mobile websocket subscription support so project lists and thread lists now refresh live from `project_id`-filtered server events instead of relying on thread-detail-only realtime updates.

### Design Rationale
- This stage prioritizes product information architecture first: users need to see projects, conversations, status, and readable history before transport-level changes like P2P.
- Keeping `session/event` as the ingestion source avoids blocking UI restructuring on bridge-side semantic capture work that is not complete yet.
- Preserving legacy routes reduces regression risk and keeps debug-oriented views available while the new primary flow stabilizes.
- Using the same websocket channel with `session_id` / `thread_id` / `project_id` filters keeps the transport surface small while letting the UI progressively shift from debug streams into conversation-oriented views.

### Notes & Caveats
- The current `thread` model is still derived from sessions, so it is presently a compatibility layer rather than a fully independent runtime source.
- Mobile thread detail now appends live thread messages, but full thread state/summary live refresh is still partial and event-shape dependent.
- Project and thread lists now refresh live, but they currently do so by refetching REST snapshots on relevant websocket events rather than consuming a dedicated thread-summary delta protocol.
- Project-page strings are not fully localized yet; this step focuses on model and navigation correctness first.
- Bridge semantic enrichment currently improves payload shape and state hints, but it does not yet provide a full real-message capture pipeline for all Codex/Claude interactions.

## [0.8.0] - 2026-03-19
### Features
- Rebased the cross-end product/design baseline from a `session/event` observer into a `project -> thread -> message` conversation model aimed at remotely viewing coding-agent progress from mobile.
- Redefined `server` at the architecture level as a P2P discovery and routing service, with `bridge <-> mobile` direct connection as the preferred privacy-preserving content path and relay only as fallback.
- Added a dedicated refactor/design document set covering the new baseline, bridge collection redesign, bridge-server protocol direction, and a unified future-change prompt for integrated implementation work.

### Design Rationale
- The previous baseline overfit the current side-channel event stream and pushed low-level transport/debug artifacts like heartbeat, raw command line, and bridge observing text into the release user experience.
- The actual user goal is to see which projects and conversations are active, what the agent said, what the user sent, whether the thread is still running or waiting, and to browse the project files from mobile.
- Making `server` the default content center would work against the privacy objective, so the new direction treats it as a registry/route broker first and a relay fallback second.

### Notes & Caveats
- This version entry records an architectural reset in documentation and planning; it does not mean the repository has already completed the full `project/thread/message` or P2P runtime implementation.
- The current runtime still primarily exposes `session/event`-oriented data and side-channel observations; follow-up implementation must explicitly distinguish `implemented`, `partially implemented`, and `not implemented`.
- `bridge` remains a standalone side-channel collector and must not be reverted to a managed launcher path as part of this redesign.

## [0.7.0] - 2026-03-17
### Features
- Refactored `bridge` from a PTY-managed agent launcher into a standalone side-channel collector with `discovery` as the default capture mode.
- Added process discovery for active `codex` / `claude` sessions plus a minimal side-channel adapter chain for `command`, `terminal_output`, `file_change`, `heartbeat`, and `error`.
- Changed default `send_prompt` handling so side-channel mode returns an explicit failed `command_result` instead of pretending prompt injection succeeded.
- Added tests for discovery classification, discovery source lifecycle, new config defaults, and side-channel command degradation.

### Design Rationale
- The new product direction requires bridge to observe independently started CLIs instead of changing the user's launch path, so discovery became the default runtime boundary.
- Kept the existing transport and protocol model stable while moving session awareness into the capture pipeline, which allows multiple discovered sessions to report through one bridge process.
- Preserved the old PTY/inbox code as optional legacy extension points rather than deleting it outright, but removed it from the default application path.

### Notes & Caveats
- The current discovery path provides a minimal side-channel implementation: it can observe active sessions and workspace changes, but it does not attach to terminal PTYs or inject prompts into standalone CLIs.
- `terminal_output` in discovery mode is currently observational bridge output, not direct CLI stdout capture.
- Legacy managed-process and inbox components still exist in the repository, but they are no longer wired into the default bridge runtime.
- Discovery filtering now prefers real CLI entry executables/scripts over broad substring matches, but it is still heuristic and may need more platform-specific refinement later.
- File watching now suppresses common generated directories and bursty duplicate writes, but it is not yet a semantic diff aggregator.

## [0.7.0] - 2026-03-17
### Features
- Connected `mobile` Session list and Log detail flows to the real server REST APIs and mobile WebSocket stream while keeping the existing mock service path available.
- Added a dedicated mobile-side server mapper so snake_case server payloads are translated into the existing Flutter MVP models without leaking transport details into pages or controllers.
- Added regression tests for real REST parsing, real WebSocket event decoding, and environment-based service bootstrapping.
- Added startup-time environment selection from `--dart-define` values so mobile can boot directly into server mode for real data debugging.
- Added a Settings-page connection test that probes `/api/health` and `/ws/mobile` before saving runtime connection values.

### Design Rationale
- Kept the real transport code inside `mobile/lib/services/real/` so the MVP module boundaries stay stable and future file-browser or prompt APIs can reuse the same service layer.
- Preserved mock-first startup because backend availability is still environment-dependent, but made the runtime `Server` switch use real implementations instead of placeholder exceptions.
- Avoided touching controller and route structure so the backend integration remains a narrow, reversible change.

### Notes & Caveats
- The app still starts in mock mode from `mobile/lib/main.dart`; switching to the real backend is currently a runtime setting and is not yet persisted.
- Real WebSocket reconnection and advanced transport state handling are still future enhancements; the current client focuses on correct subscription and event decoding.
- The Settings connection test validates reachability and handshake only; it does not yet inspect session payload quality or live-stream semantics beyond successful connect.

## [0.6.0] - 2026-03-17
### Features
- Added `bridge/cmd/fake-source` and `server/cmd/debug-subscriber` so the bridge and server can be integration-tested without `mobile`.
- Added `jsonl` structured input mode to the bridge so scripted event streams can drive `terminal_output`, `ai_output`, `file_change`, and `heartbeat` reporting through the real bridge transport.
- Updated the server event path to auto-create sessions from incoming bridge metadata, persist incoming events, and keep REST lookup plus WebSocket broadcast on the same ingestion flow.
- Added regression coverage for bridge JSONL input, server auto-session creation, router-level broadcast, inbox state persistence, and managed-process restart/error reporting.

### Design Rationale
- Kept the new test loop inside the existing module boundaries so `fake-source` exercises the real bridge runtime instead of bypassing it with ad-hoc scripts.
- Let the server create sessions on first contact because the integration loop needs one less manual bootstrap step and the session metadata is already present in bridge payloads.
- Preserved the existing prompt and managed-process skeleton, but broke the internal package cycle by introducing a neutral bridge message shape for command transport.

### Notes & Caveats
- Session and event persistence are still memory-backed in the server, so process restarts clear history.
- The repository is a multi-module workspace; verification is run as `go test ./...` and `go build ./...` inside `bridge/` and `server/`, not from the workspace root.

## [0.1.0] - 2026-03-17
### Features
- Created the initial multi-project workspace layout for `server`, `bridge`, and `mobile`.
- Added a compilable Go server scaffold with routing and a health endpoint.
- Added a compilable Go bridge scaffold with configuration and runtime startup flow.
- Added a Flutter-oriented mobile directory structure with placeholder Dart files and module layout.

### Design Rationale
- Split the codebase by runtime boundary so each executable can evolve independently.
- Kept the first pass lightweight and explicit, prioritizing a clean handoff into real implementation work.
- Used separate Go modules with a root `go.work` file so the backend components remain isolated without fragmenting the repository.

### Notes & Caveats
- The mobile client is not a generated Flutter project yet because the Flutter SDK is not available in the current environment.
- The server and bridge are scaffolds only; business logic, protocol handling, and persistence are still pending.

## [0.2.0] - 2026-03-17
### Features
- Implemented the `bridge` MVP event pipeline with session-aware message envelopes for `terminal_output`, `ai_output`, `command`, `file_change`, `heartbeat`, and `error`.
- Added a reconnecting WebSocket transport with heartbeat emission, inbound command handling, and buffered outbound delivery.
- Added stdin/text-stream capture and a minimal `send_prompt` command acknowledgement flow.
- Added bridge-focused tests for config loading, message encoding, capture behavior, and transport reconnect/heartbeat behavior.
- Added `bridge/README.md` with environment variables, run instructions, and protocol examples.

### Design Rationale
- Kept the MVP centered on a single clear path: local text capture to WebSocket reporting, with lightweight command round-trips.
- Moved protocol and session metadata into `internal/session` so capture, transport, and command handling share one message model.
- Used a replaceable capture source instead of a PTY hook first, so the runtime path is usable now without locking the project into terminal-specific behavior too early.

### Notes & Caveats
- The current bridge captures standard input as a text stream; real PTY interception and file watcher integration are still future work.
- `send_prompt` is acknowledged and observable, but it is not yet injected into a real AI agent process.
- The server-side protocol is now explicit, but persistence, auth, and full session recovery still depend on future server implementation.

## [0.2.0] - 2026-03-17
### Features
- Implemented the server MVP main loop for session management, bridge event ingestion, in-memory persistence, and mobile real-time subscription.
- Added REST endpoints for session creation, listing, detail lookup, status updates, and per-session event history.
- Added separate bridge and mobile WebSocket endpoints with message validation, event broadcast, and subscriber filtering by `session_id`.
- Added store, protocol, router, and WebSocket integration tests to lock in the MVP behavior.

### Design Rationale
- Kept the architecture layered as `handler -> service -> store` so Gin handlers stay thin and SQLite can be added behind interfaces later.
- Used in-memory stores for MVP speed while preserving explicit `SessionStore` and `EventStore` boundaries for future persistence work.
- Centralized fan-out in an event hub so bridge ingestion and mobile subscription remain decoupled from each other.

### Notes & Caveats
- Persistence is process-local only in this version; restarting the server clears sessions and event history.
- WebSocket auth, prompt delivery, and file browser APIs are still intentionally out of scope for this MVP.

## [0.3.0] - 2026-03-17
### Features
- Added server-side file browsing APIs with workspace-root confinement, ignore rules, preview whitelisting, and explicit non-previewable responses for binary or oversized files.
- Added prompt library APIs plus prompt command dispatch from server to bridge, with command task tracking and command-result ingestion.
- Upgraded the bridge managed-process path to execute prompts against the local PTY-backed agent and return captured output as the command result.
- Added persistent prompt inbox consumer state in bridge to avoid replaying previously consumed prompt records after restart.

### Design Rationale
- Kept file browsing isolated in a dedicated service so path validation and preview policy are enforced consistently and do not leak into HTTP handlers.
- Routed prompt delivery through explicit command tasks and a bridge registry so server-side control remains observable and extensible for future approval, retry, or persistence features.
- Used direct PTY prompt execution only when a managed local agent is configured; this preserves a lightweight fallback path while enabling true end-to-end prompt execution in the primary bridge mode.

### Notes & Caveats
- File preview remains intentionally limited to a text-oriented whitelist and a size cap; unsupported files are visible in the tree but not rendered.
- Prompt execution results are derived from managed-process output observed after prompt injection, so long-running agents may require future protocol-level streaming or completion markers for stronger result boundaries.

## [0.3.0] - 2026-03-17
### Features
- Added recursive workspace file watching to the bridge and now emit `file_change` events with relative paths and operation types.
- Added local prompt injection for `send_prompt` by writing JSON lines into a configurable prompt inbox file.
- Extended bridge tests to cover file watcher behavior and command-to-local-inbox execution.

### Design Rationale
- Used a file-based prompt sink as the first local injection target because it is deterministic, observable, and easy to replace with a real CLI/stdin adapter later.
- Kept file watching scoped to the workspace root with ignore rules for `.git`, `node_modules`, and `.codescope` to avoid noisy or self-induced events.

### Notes & Caveats
- File watching currently reports raw create/write/remove/rename/chmod events and does not deduplicate bursty editor writes.
- Prompt injection is local persistence only; a downstream agent process must still consume the inbox file.

## [0.4.0] - 2026-03-17
### Features
- Added an inbox consumer that tails the local prompt inbox and forwards new prompts to a configured local target.
- Added optional PTY-managed agent execution so the bridge can launch a local command, capture terminal output, and inject prompts through PTY stdin.
- Added configuration for managed command and arguments, plus tests for inbox consumption and managed-process prompt delivery.

### Design Rationale
- Kept prompt delivery file-backed so remote command handling remains observable and replayable even before direct server-to-process injection is fully trusted.
- Used one managed-process abstraction for both capture and prompt delivery to avoid splitting terminal output and command input across unrelated code paths.

### Notes & Caveats
- `CODESCOPE_BRIDGE_MANAGED_ARGS` is currently parsed with simple whitespace splitting and does not support shell-style quoting.
- PTY behavior depends on local platform support and the target CLI’s ability to run interactively over a pseudo-terminal.

## [0.5.0] - 2026-03-17
### Features
- Added persistent prompt inbox offsets so the consumer resumes from the last processed byte after bridge restarts.
- Added managed-agent restart policy controls for max retries and restart delay.
- Added managed-process lifecycle error events that report exit message, derived exit status, restart attempt, and whether another restart is scheduled.
- Added coverage for offset persistence, replay avoidance, and managed-process restart behavior.

### Design Rationale
- Kept consumer state byte-offset based because the inbox is append-only JSONL and offset persistence is the smallest reliable dedupe mechanism.
- Emitted managed-process failures as protocol `error` events so server-side monitoring can stay within the existing MVP event model instead of introducing a parallel lifecycle channel.

### Notes & Caveats
- Offset persistence assumes the inbox remains append-only; if external tools rewrite lines in place, replay guarantees do not hold.
- Restart policy currently applies uniformly to startup failures and runtime exits; there is no differentiated backoff by error category yet.

## [0.6.0] - 2026-03-17
### Features
- Hardened prompt inbox recovery by persisting the last consumed line fingerprint alongside the byte offset and resetting safely when the inbox is rewritten before the stored offset.
- Added failure-stage classification for managed agents (`startup`, `runtime`, `exit`) and now emit restart delay metadata in error events.
- Added differentiated restart backoff by failure stage instead of using one uniform delay for all managed-process failures.

### Design Rationale
- Using the last consumed line hash is a pragmatic middle ground: it catches truncation/rewrite cases that a raw byte offset misses without turning the inbox consumer into a database.
- Splitting managed-process failures into startup/runtime/exit keeps monitoring signals actionable and allows backoff to match the failure mode instead of treating all exits as equivalent.

### Notes & Caveats
- The inbox rewrite check validates only the last consumed record boundary; it is optimized for append-only JSONL and not for arbitrary in-place edits across the whole file.
- Restart backoff is stage-aware but still policy-based, not adaptive; it does not yet inspect specific exit codes or stderr signatures.
