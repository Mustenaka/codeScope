import 'dart:async';

import 'package:codescope_mobile/modules/file/file_content_record.dart';
import 'package:codescope_mobile/modules/file/file_tree_node.dart';
import 'package:codescope_mobile/modules/log/log_event.dart';
import 'package:codescope_mobile/modules/message/thread_message_record.dart';
import 'package:codescope_mobile/modules/project/project_record.dart';
import 'package:codescope_mobile/modules/prompt/prompt_command_task.dart';
import 'package:codescope_mobile/modules/prompt/prompt_controller.dart';
import 'package:codescope_mobile/modules/session/session_record.dart';
import 'package:codescope_mobile/modules/thread/thread_record.dart';
import 'package:codescope_mobile/services/rest_client.dart';
import 'package:codescope_mobile/services/websocket_client.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('loads tasks and updates matching task when command result arrives',
      () async {
    final socketEvents = StreamController<LogEvent>();
    final controller = PromptController(
      threadId: 'thread-001',
      restClient: _FakePromptRestClient(),
      webSocketClient: _FakePromptWebSocketClient(socketEvents.stream),
    );

    await controller.load();
    expect(controller.tasks.single.status, PromptCommandTaskStatus.running);

    socketEvents.add(
      LogEvent(
        id: 'event-002',
        sessionId: 'session-001',
        messageType: 'command_result',
        type: LogEventType.commandResult,
        level: LogLevel.warning,
        content: 'side-channel mode does not support prompt injection',
        createdAt: DateTime.parse('2026-03-18T09:02:00Z'),
        metadata: const <String, Object?>{
          'accepted': false,
          'error': 'side-channel mode does not support prompt injection',
        },
        commandId: 'cmd-001',
        commandType: 'send_prompt',
        commandStatus: 'failed',
      ),
    );

    await Future<void>.delayed(Duration.zero);

    expect(controller.tasks.single.status, PromptCommandTaskStatus.failed);
    expect(
      controller.tasks.single.result,
      'side-channel mode does not support prompt injection',
    );

    await controller.disposeAsync();
    await socketEvents.close();
  });

  test('sends prompt and prepends created task', () async {
    final restClient = _FakePromptRestClient();
    final controller = PromptController(
      threadId: 'thread-001',
      restClient: restClient,
      webSocketClient:
          _FakePromptWebSocketClient(const Stream<LogEvent>.empty()),
    );

    await controller.load();
    await controller.sendPrompt('continue implementing mobile prompt flow');

    expect(restClient.lastPrompt, 'continue implementing mobile prompt flow');
    expect(controller.tasks.first.id, 'cmd-002');
    expect(controller.tasks.first.prompt,
        'continue implementing mobile prompt flow');
    await controller.disposeAsync();
  });
}

class _FakePromptRestClient implements CodeScopeRestClient {
  String? lastPrompt;

  @override
  Future<List<ProjectRecord>> fetchProjects() {
    throw UnimplementedError();
  }

  @override
  Future<ProjectRecord> fetchProjectDetail(String projectId) {
    throw UnimplementedError();
  }

  @override
  Future<List<ThreadRecord>> fetchProjectThreads(String projectId) {
    throw UnimplementedError();
  }

  @override
  Future<ThreadRecord> createProjectThread(String projectId, String content) {
    throw UnimplementedError();
  }

  @override
  Future<ThreadRecord> fetchThreadDetail(String threadId) {
    return Future<ThreadRecord>.value(
      ThreadRecord(
        id: 'thread-001',
        projectId: 'project-001',
        sessionId: 'session-001',
        title: 'Release thread',
        status: ThreadStatus.running,
        summary: 'continue from last step',
        lastActivityAt: DateTime.parse('2026-03-18T09:01:00Z'),
        startedAt: DateTime.parse('2026-03-18T09:00:00Z'),
      ),
    );
  }

  @override
  Future<List<ThreadMessageRecord>> fetchThreadMessages(String threadId) {
    throw UnimplementedError();
  }

  @override
  Future<List<PromptCommandTask>> fetchThreadCommands(String threadId) async {
    return fetchSessionCommands('session-001');
  }

  @override
  Future<PromptCommandTask> sendThreadPrompt(String threadId, String content) {
    return sendPrompt('session-001', content);
  }

  @override
  Future<List<FileTreeNode>> fetchProjectFileTree(String projectId) {
    throw UnimplementedError();
  }

  @override
  Future<FileContentRecord> fetchProjectFileContent(
    String projectId,
    String path,
  ) {
    throw UnimplementedError();
  }

  @override
  Future<List<PromptCommandTask>> fetchSessionCommands(String sessionId) async {
    return <PromptCommandTask>[
      PromptCommandTask(
        id: 'cmd-001',
        sessionId: sessionId,
        taskType: 'send_prompt',
        prompt: 'continue from last step',
        status: PromptCommandTaskStatus.running,
        result: '',
        createdAt: DateTime.parse('2026-03-18T09:00:00Z'),
        updatedAt: DateTime.parse('2026-03-18T09:01:00Z'),
      ),
    ];
  }

  @override
  Future<PromptCommandTask> sendPrompt(String sessionId, String content) async {
    lastPrompt = content;
    return PromptCommandTask(
      id: 'cmd-002',
      sessionId: sessionId,
      taskType: 'send_prompt',
      prompt: content,
      status: PromptCommandTaskStatus.running,
      result: '',
      createdAt: DateTime.parse('2026-03-18T09:03:00Z'),
      updatedAt: DateTime.parse('2026-03-18T09:03:00Z'),
    );
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
  Future<List<LogEvent>> fetchSessionEvents(String sessionId) async {
    return <LogEvent>[
      LogEvent(
        id: 'event-001',
        sessionId: sessionId,
        messageType: 'event',
        type: LogEventType.command,
        level: LogLevel.info,
        content: 'continue from last step',
        createdAt: DateTime.parse('2026-03-18T09:00:00Z'),
        metadata: const <String, Object?>{
          'content': 'continue from last step',
        },
        commandId: 'cmd-001',
        commandType: 'send_prompt',
      ),
    ];
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

class _FakePromptWebSocketClient implements CodeScopeWebSocketClient {
  _FakePromptWebSocketClient(this.stream);

  final Stream<LogEvent> stream;

  @override
  Stream<LogEvent> subscribeToProjects() => const Stream.empty();

  @override
  Stream<LogEvent> subscribeToProject(String projectId) => const Stream.empty();

  @override
  Stream<LogEvent> subscribeToSession(String sessionId) => stream;

  @override
  Stream<ThreadMessageRecord> subscribeToThread(String threadId) =>
      const Stream.empty();
}
