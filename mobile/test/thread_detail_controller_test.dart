import 'dart:async';

import 'package:codescope_mobile/modules/file/file_content_record.dart';
import 'package:codescope_mobile/modules/file/file_tree_node.dart';
import 'package:codescope_mobile/modules/log/log_event.dart';
import 'package:codescope_mobile/modules/message/thread_detail_controller.dart';
import 'package:codescope_mobile/modules/message/thread_message_record.dart';
import 'package:codescope_mobile/modules/project/project_record.dart';
import 'package:codescope_mobile/modules/prompt/prompt_command_task.dart';
import 'package:codescope_mobile/modules/session/session_record.dart';
import 'package:codescope_mobile/modules/thread/thread_record.dart';
import 'package:codescope_mobile/services/rest_client.dart';
import 'package:codescope_mobile/services/websocket_client.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('loads thread snapshot and appends live thread messages', () async {
    final liveMessages = StreamController<ThreadMessageRecord>();
    final controller = ThreadDetailController(
      _FakeRestClient(),
      _FakeWebSocketClient(liveMessages.stream),
    );

    await controller.load('thread-001');
    expect(controller.messages, hasLength(1));

    liveMessages.add(
      ThreadMessageRecord(
        id: 'message-002',
        threadId: 'thread-001',
        role: ThreadMessageRole.assistant,
        content: 'Applied follow-up patch',
        createdAt: DateTime.parse('2026-03-19T10:06:00Z'),
        sequence: 2,
        sourceType: 'ai_output',
      ),
    );
    await Future<void>.delayed(Duration.zero);

    expect(controller.messages, hasLength(2));
    expect(controller.messages.last.content, 'Applied follow-up patch');

    await liveMessages.close();
    controller.dispose();
  });
}

class _FakeRestClient implements CodeScopeRestClient {
  @override
  Future<ThreadRecord> fetchThreadDetail(String threadId) async {
    return ThreadRecord(
      id: threadId,
      projectId: 'project-001',
      sessionId: 'session-001',
      title: 'Implemented server API',
      status: ThreadStatus.running,
      summary: 'Implemented server API',
      lastActivityAt: DateTime.parse('2026-03-19T10:05:00Z'),
      startedAt: DateTime.parse('2026-03-19T10:00:00Z'),
    );
  }

  @override
  Future<List<ThreadMessageRecord>> fetchThreadMessages(String threadId) async {
    return <ThreadMessageRecord>[
      ThreadMessageRecord(
        id: 'message-001',
        threadId: threadId,
        role: ThreadMessageRole.user,
        content: 'continue implementing',
        createdAt: DateTime.parse('2026-03-19T10:01:00Z'),
        sequence: 1,
        sourceType: 'command',
      ),
    ];
  }

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

  final Stream<ThreadMessageRecord> stream;

  @override
  Stream<LogEvent> subscribeToProjects() => const Stream.empty();

  @override
  Stream<LogEvent> subscribeToProject(String projectId) => const Stream.empty();

  @override
  Stream<LogEvent> subscribeToSession(String sessionId) => const Stream.empty();

  @override
  Stream<ThreadMessageRecord> subscribeToThread(String threadId) => stream;
}
