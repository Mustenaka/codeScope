import 'dart:async';

import 'package:codescope_mobile/modules/file/file_content_record.dart';
import 'package:codescope_mobile/modules/file/file_tree_node.dart';
import 'package:codescope_mobile/modules/log/log_event.dart';
import 'package:codescope_mobile/modules/message/thread_message_record.dart';
import 'package:codescope_mobile/modules/project/project_record.dart';
import 'package:codescope_mobile/modules/prompt/prompt_command_task.dart';
import 'package:codescope_mobile/modules/session/session_record.dart';
import 'package:codescope_mobile/modules/thread/thread_list_controller.dart';
import 'package:codescope_mobile/modules/thread/thread_record.dart';
import 'package:codescope_mobile/services/rest_client.dart';
import 'package:codescope_mobile/services/websocket_client.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('refreshes thread list after a live project event', () async {
    final streamController = StreamController<LogEvent>();
    final restClient = _MutableThreadRestClient();
    final controller = ThreadListController(
      restClient,
      _FakeWebSocketClient(streamController.stream),
    );

    await controller.loadThreads('project-1');
    await controller.startRealtime('project-1');
    expect(controller.threads, hasLength(1));

    restClient.threads = <ThreadRecord>[
      ThreadRecord(
        id: 'thread-1',
        projectId: 'project-1',
        sessionId: 'session-1',
        title: 'Implemented server API',
        status: ThreadStatus.running,
        summary: 'Implemented server API',
        lastActivityAt: DateTime.parse('2026-03-19T10:05:00Z'),
        startedAt: DateTime.parse('2026-03-19T10:00:00Z'),
      ),
      ThreadRecord(
        id: 'thread-2',
        projectId: 'project-1',
        sessionId: 'session-2',
        title: 'Waiting for next prompt',
        status: ThreadStatus.waitingPrompt,
        summary: 'Awaiting input',
        lastActivityAt: DateTime.parse('2026-03-19T10:06:00Z'),
        startedAt: DateTime.parse('2026-03-19T10:02:00Z'),
      ),
    ];
    streamController.add(
      LogEvent(
        id: 'event-live-2',
        sessionId: 'session-2',
        messageType: 'event',
        type: LogEventType.aiOutput,
        level: LogLevel.info,
        content: 'Waiting for input',
        createdAt: DateTime.parse('2026-03-19T10:06:00Z'),
        metadata: const <String, Object?>{'project_id': 'project-1'},
      ),
    );
    await Future<void>.delayed(const Duration(milliseconds: 20));

    expect(controller.threads, hasLength(2));
    await streamController.close();
    controller.dispose();
  });
}

class _MutableThreadRestClient implements CodeScopeRestClient {
  List<ThreadRecord> threads = <ThreadRecord>[
    ThreadRecord(
      id: 'thread-1',
      projectId: 'project-1',
      sessionId: 'session-1',
      title: 'Implemented server API',
      status: ThreadStatus.running,
      summary: 'Implemented server API',
      lastActivityAt: DateTime.parse('2026-03-19T10:05:00Z'),
      startedAt: DateTime.parse('2026-03-19T10:00:00Z'),
    ),
  ];

  @override
  Future<List<ThreadRecord>> fetchProjectThreads(String projectId) async => threads;

  @override
  Future<List<ProjectRecord>> fetchProjects() {
    throw UnimplementedError();
  }

  @override
  Future<ProjectRecord> fetchProjectDetail(String projectId) {
    throw UnimplementedError();
  }

  @override
  Future<ThreadRecord> fetchThreadDetail(String threadId) {
    throw UnimplementedError();
  }

  @override
  Future<List<ThreadMessageRecord>> fetchThreadMessages(String threadId) {
    throw UnimplementedError();
  }

  @override
  Future<List<SessionRecord>> fetchSessions() {
    throw UnimplementedError();
  }

  @override
  Future<SessionRecord> fetchSessionDetail(String sessionId) {
    throw UnimplementedError();
  }

  @override
  Future<List<LogEvent>> fetchSessionEvents(String sessionId) {
    throw UnimplementedError();
  }

  @override
  Future<List<PromptCommandTask>> fetchSessionCommands(String sessionId) {
    throw UnimplementedError();
  }

  @override
  Future<PromptCommandTask> sendPrompt(String sessionId, String content) {
    throw UnimplementedError();
  }

  @override
  Future<List<FileTreeNode>> fetchSessionFileTree(String sessionId) {
    throw UnimplementedError();
  }

  @override
  Future<FileContentRecord> fetchSessionFileContent(
    String sessionId,
    String path,
  ) {
    throw UnimplementedError();
  }
}

class _FakeWebSocketClient implements CodeScopeWebSocketClient {
  _FakeWebSocketClient(this.stream);

  final Stream<LogEvent> stream;

  @override
  Stream<LogEvent> subscribeToProjects() => stream;

  @override
  Stream<LogEvent> subscribeToProject(String projectId) => stream;

  @override
  Stream<LogEvent> subscribeToSession(String sessionId) => stream;

  @override
  Stream<ThreadMessageRecord> subscribeToThread(String threadId) =>
      const Stream.empty();
}
