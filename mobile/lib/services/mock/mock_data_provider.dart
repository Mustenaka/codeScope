import 'dart:async';
import 'dart:math';

import '../../modules/file/file_content_record.dart';
import '../../modules/file/file_tree_node.dart';
import '../../modules/log/log_event.dart';
import '../../modules/message/thread_message_record.dart';
import '../../modules/project/project_record.dart';
import '../../modules/prompt/prompt_command_task.dart';
import '../../modules/session/session_record.dart';
import '../../modules/thread/thread_record.dart';

class MockDataProvider {
  MockDataProvider() : _startedAt = DateTime.now().toUtc();

  final DateTime _startedAt;
  final Random _random = Random(7);

  List<SessionRecord> buildSessions() {
    return <SessionRecord>[
      SessionRecord(
        id: 'session-001',
        projectName: 'codeScope/server',
        workspaceRoot: '/workspace/codeScope',
        machineId: 'mbp-mumte',
        status: SessionStatus.running,
        startedAt: _startedAt.subtract(const Duration(minutes: 18)),
        updatedAt: _startedAt.subtract(const Duration(seconds: 5)),
        summary: 'Bridge connected, streaming terminal and AI output.',
      ),
      SessionRecord(
        id: 'session-002',
        projectName: 'codeScope/mobile',
        workspaceRoot: '/workspace/codeScope/mobile',
        machineId: 'win-devbox',
        status: SessionStatus.failed,
        startedAt: _startedAt.subtract(const Duration(hours: 3, minutes: 12)),
        endedAt: _startedAt.subtract(const Duration(hours: 3)),
        updatedAt: _startedAt.subtract(const Duration(hours: 3)),
        summary: 'Flutter test run stopped on missing SDK.',
      ),
      SessionRecord(
        id: 'session-003',
        projectName: 'internal/agent-bridge',
        workspaceRoot: '/workspace/agent-bridge',
        machineId: 'linux-builder',
        status: SessionStatus.stopped,
        startedAt: _startedAt.subtract(const Duration(days: 1, hours: 1)),
        endedAt: _startedAt.subtract(const Duration(days: 1)),
        updatedAt: _startedAt.subtract(const Duration(days: 1)),
        summary: 'No active stream, last sync completed normally.',
      ),
    ];
  }

  List<ProjectRecord> buildProjects() {
    return <ProjectRecord>[
      ProjectRecord(
        id: 'project-001',
        name: 'codeScope',
        workspaceRoot: '/workspace/codeScope',
        machineId: 'mbp-mumte',
        threadCount: 2,
        runningThreadCount: 1,
        lastActivityAt: _startedAt.subtract(const Duration(seconds: 5)),
      ),
      ProjectRecord(
        id: 'project-002',
        name: 'agent-bridge',
        workspaceRoot: '/workspace/agent-bridge',
        machineId: 'linux-builder',
        threadCount: 1,
        runningThreadCount: 0,
        lastActivityAt: _startedAt.subtract(const Duration(days: 1)),
      ),
    ];
  }

  ProjectRecord buildProjectDetail(String projectId) {
    return buildProjects().firstWhere((project) => project.id == projectId);
  }

  List<ThreadRecord> buildThreads(String projectId) {
    switch (projectId) {
      case 'project-001':
        return <ThreadRecord>[
          ThreadRecord(
            id: 'thread-001',
            projectId: 'project-001',
            sessionId: 'session-001',
            title: 'Implemented server API',
            agentKind: 'codex',
            status: ThreadStatus.running,
            summary: 'Implemented server API',
            lastActivityAt: _startedAt.subtract(const Duration(seconds: 5)),
            startedAt: _startedAt.subtract(const Duration(minutes: 18)),
          ),
          ThreadRecord(
            id: 'thread-002',
            projectId: 'project-001',
            sessionId: 'session-002',
            title: 'Waiting for next prompt',
            agentKind: 'codex',
            status: ThreadStatus.waitingPrompt,
            summary: 'Prompt injection unavailable in side-channel mode.',
            lastActivityAt: _startedAt.subtract(const Duration(minutes: 16)),
            startedAt: _startedAt.subtract(const Duration(minutes: 30)),
          ),
        ];
      default:
        return <ThreadRecord>[
          ThreadRecord(
            id: 'thread-003',
            projectId: 'project-002',
            sessionId: 'session-003',
            title: 'Last sync completed',
            agentKind: 'claude',
            status: ThreadStatus.completed,
            summary: 'No active stream, last sync completed normally.',
            lastActivityAt: _startedAt.subtract(const Duration(days: 1)),
            startedAt: _startedAt.subtract(const Duration(days: 1, hours: 1)),
            endedAt: _startedAt.subtract(const Duration(days: 1)),
          ),
        ];
    }
  }

  ThreadRecord buildThreadDetail(String threadId) {
    return <ThreadRecord>[
      ...buildThreads('project-001'),
      ...buildThreads('project-002'),
    ].firstWhere((thread) => thread.id == threadId);
  }

  List<ThreadMessageRecord> buildThreadMessages(String threadId) {
    switch (threadId) {
      case 'thread-001':
        return <ThreadMessageRecord>[
          ThreadMessageRecord(
            id: 'message-001',
            threadId: threadId,
            role: ThreadMessageRole.user,
            content: 'continue implementing',
            createdAt: _startedAt.subtract(const Duration(minutes: 17)),
            sequence: 1,
            sourceType: 'command',
          ),
          ThreadMessageRecord(
            id: 'message-002',
            threadId: threadId,
            role: ThreadMessageRole.assistant,
            content: 'Implemented server API',
            createdAt: _startedAt.subtract(const Duration(seconds: 5)),
            sequence: 2,
            sourceType: 'ai_output',
          ),
        ];
      default:
        return <ThreadMessageRecord>[
          ThreadMessageRecord(
            id: 'message-003',
            threadId: threadId,
            role: ThreadMessageRole.system,
            content: 'No message history available yet.',
            createdAt: _startedAt.subtract(const Duration(days: 1)),
            sequence: 1,
            sourceType: 'system',
          ),
        ];
    }
  }

  SessionRecord buildSessionDetail(String sessionId) {
    return buildSessions().firstWhere((session) => session.id == sessionId);
  }

  List<LogEvent> buildInitialEvents(String sessionId) {
    final now = _startedAt;
    return <LogEvent>[
      LogEvent(
        id: '$sessionId-event-001',
        sessionId: sessionId,
        messageType: 'event',
        type: LogEventType.command,
        level: LogLevel.info,
        content: 'codex exec --task "sync mobile MVP"',
        createdAt: now.subtract(const Duration(minutes: 18)),
        metadata: const {
          'stream': 'stdin',
          'content': 'codex exec --task "sync mobile MVP"',
          'source': 'process_discovery',
        },
        commandId: 'cmd-observed-001',
      ),
      LogEvent(
        id: '$sessionId-event-002',
        sessionId: sessionId,
        messageType: 'event',
        type: LogEventType.terminalOutput,
        level: LogLevel.info,
        content: 'Scanning repo structure and service boundaries...',
        createdAt: now.subtract(const Duration(minutes: 17, seconds: 40)),
        metadata: const {'stream': 'stdout'},
      ),
      LogEvent(
        id: '$sessionId-event-003',
        sessionId: sessionId,
        messageType: 'event',
        type: LogEventType.aiOutput,
        level: LogLevel.info,
        content: 'Prepared session list and live log page skeleton.',
        createdAt: now.subtract(const Duration(minutes: 17)),
        metadata: const {'agent': 'codex'},
      ),
      LogEvent(
        id: '$sessionId-event-003b',
        sessionId: sessionId,
        messageType: 'command_result',
        type: LogEventType.commandResult,
        level: LogLevel.warning,
        content: 'side-channel mode does not support prompt injection',
        createdAt: now.subtract(const Duration(minutes: 16, seconds: 45)),
        metadata: const {
          'accepted': false,
          'error': 'side-channel mode does not support prompt injection',
        },
        commandId: 'cmd-observed-001',
        commandType: 'send_prompt',
        commandStatus: 'failed',
      ),
      LogEvent(
        id: '$sessionId-event-004',
        sessionId: sessionId,
        messageType: 'event',
        type: LogEventType.fileChange,
        level: LogLevel.info,
        content: 'Updated lib/modules/session/session_page.dart',
        createdAt: now.subtract(const Duration(minutes: 16, seconds: 30)),
        metadata: const {
          'path': 'lib/modules/session/session_page.dart',
          'op': 'write',
        },
      ),
      LogEvent(
        id: '$sessionId-event-005',
        sessionId: sessionId,
        messageType: 'event',
        type: LogEventType.error,
        level: LogLevel.warning,
        content:
            'Flutter CLI unavailable in current environment, mock mode enabled.',
        createdAt: now.subtract(const Duration(minutes: 16)),
        metadata: const {'retryable': true},
      ),
    ];
  }

  Stream<LogEvent> buildLiveEvents(String sessionId) async* {
    final eventTypes = <LogEventType>[
      LogEventType.heartbeat,
      LogEventType.terminalOutput,
      LogEventType.aiOutput,
    ];

    var index = 0;
    while (true) {
      await Future<void>.delayed(const Duration(seconds: 2));
      final type = eventTypes[index % eventTypes.length];
      final createdAt = DateTime.now().toUtc();
      index += 1;

      yield LogEvent(
        id: '$sessionId-live-$index',
        sessionId: sessionId,
        messageType: type == LogEventType.heartbeat ? 'heartbeat' : 'event',
        type: type,
        level: type == LogEventType.heartbeat ? LogLevel.debug : LogLevel.info,
        content: _buildLiveContent(type, index),
        createdAt: createdAt,
        metadata: <String, Object?>{
          'latencyMs': 120 + _random.nextInt(80),
          'sequence': index,
        },
      );
    }
  }

  String _buildLiveContent(LogEventType type, int index) {
    switch (type) {
      case LogEventType.heartbeat:
        return 'Bridge heartbeat acknowledged.';
      case LogEventType.terminalOutput:
        return 'Tail chunk #$index received from remote terminal.';
      case LogEventType.aiOutput:
        return 'Agent checkpoint #$index completed; waiting for next task.';
      case LogEventType.commandResult:
        return 'Agent command result received.';
      case LogEventType.command:
      case LogEventType.fileChange:
      case LogEventType.diff:
      case LogEventType.error:
        return 'Unsupported mock event.';
    }
  }

  List<FileTreeNode> buildFileTree(String sessionId) {
    return const <FileTreeNode>[
      FileTreeNode(
        name: 'lib',
        path: 'lib',
        type: 'directory',
        children: <FileTreeNode>[
          FileTreeNode(
            name: 'main.dart',
            path: 'lib/main.dart',
            type: 'file',
            size: 128,
            previewable: true,
          ),
        ],
      ),
      FileTreeNode(
        name: 'README.md',
        path: 'README.md',
        type: 'file',
        size: 256,
        previewable: true,
      ),
    ];
  }

  FileContentRecord buildFileContent(String sessionId, String path) {
    switch (path) {
      case 'lib/main.dart':
        return const FileContentRecord(
          path: 'lib/main.dart',
          size: 128,
          previewable: true,
          language: 'dart',
          content: 'void main() {\n  runApp(const Placeholder());\n}',
        );
      default:
        return FileContentRecord(
          path: path,
          size: 0,
          previewable: false,
          reason: 'extension_not_previewable',
        );
    }
  }

  List<PromptCommandTask> buildPromptCommands(String sessionId) {
    return <PromptCommandTask>[
      PromptCommandTask(
        id: 'cmd-001',
        sessionId: sessionId,
        taskType: 'send_prompt',
        prompt: 'continue from last step',
        status: PromptCommandTaskStatus.failed,
        result: 'side-channel mode does not support prompt injection',
        createdAt: _startedAt.subtract(
          const Duration(minutes: 16, seconds: 50),
        ),
        updatedAt: _startedAt.subtract(
          const Duration(minutes: 16, seconds: 45),
        ),
      ),
    ];
  }

  PromptCommandTask createPromptCommand(String sessionId, String content) {
    final now = DateTime.now().toUtc();
    return PromptCommandTask(
      id: 'cmd-${now.microsecondsSinceEpoch}',
      sessionId: sessionId,
      taskType: 'send_prompt',
      prompt: content,
      status: PromptCommandTaskStatus.running,
      result: '',
      createdAt: now,
      updatedAt: now,
    );
  }
}
