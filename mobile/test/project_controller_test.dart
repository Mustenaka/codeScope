import 'dart:async';

import 'package:codescope_mobile/modules/file/file_content_record.dart';
import 'package:codescope_mobile/modules/file/file_tree_node.dart';
import 'package:codescope_mobile/modules/log/log_event.dart';
import 'package:codescope_mobile/modules/message/thread_message_record.dart';
import 'package:codescope_mobile/modules/project/project_controller.dart';
import 'package:codescope_mobile/modules/project/project_record.dart';
import 'package:codescope_mobile/modules/prompt/prompt_command_task.dart';
import 'package:codescope_mobile/modules/session/session_record.dart';
import 'package:codescope_mobile/modules/thread/thread_record.dart';
import 'package:codescope_mobile/services/rest_client.dart';
import 'package:codescope_mobile/services/websocket_client.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('loads projects and sorts by last activity descending', () async {
    final controller = ProjectController(
      _FakeRestClient(<ProjectRecord>[
        ProjectRecord(
          id: 'project-1',
          name: 'codeScope',
          workspaceRoot: 'D:/Work/Code/Cross/codeScope',
          machineId: 'devbox',
          threadCount: 2,
          runningThreadCount: 1,
          createdAt: DateTime.parse('2026-03-19T09:45:00Z'),
          lastActivityAt: DateTime.parse('2026-03-19T10:05:00Z'),
        ),
        ProjectRecord(
          id: 'project-2',
          name: 'other',
          workspaceRoot: 'D:/Work/Code/Cross/other',
          machineId: 'devbox',
          threadCount: 1,
          runningThreadCount: 0,
          createdAt: DateTime.parse('2026-03-19T08:45:00Z'),
          lastActivityAt: DateTime.parse('2026-03-19T09:05:00Z'),
        ),
      ]),
      _FakeWebSocketClient(const Stream<LogEvent>.empty()),
    );

    await controller.loadProjects();

    expect(controller.projects, hasLength(2));
    expect(controller.projects.first.name, 'codeScope');
    expect(controller.projects.first.threadCount, 2);
    controller.dispose();
  });

  test('refreshes projects after a live project event', () async {
    final streamController = StreamController<LogEvent>();
    final restClient = _MutableProjectRestClient();
    final controller = ProjectController(
      restClient,
      _FakeWebSocketClient(streamController.stream),
    );

    await controller.loadProjects();
    await controller.startRealtime();
    expect(controller.projects.first.threadCount, 1);

    restClient.projects = <ProjectRecord>[
      ProjectRecord(
        id: 'project-1',
        name: 'codeScope',
        workspaceRoot: 'D:/Work/Code/Cross/codeScope',
        machineId: 'devbox',
        threadCount: 2,
        runningThreadCount: 1,
        createdAt: DateTime.parse('2026-03-19T09:45:00Z'),
        lastActivityAt: DateTime.parse('2026-03-19T10:06:00Z'),
      ),
    ];
    streamController.add(
      LogEvent(
        id: 'event-live-1',
        sessionId: 'session-1',
        messageType: 'event',
        type: LogEventType.aiOutput,
        level: LogLevel.info,
        content: 'Applied live patch',
        createdAt: DateTime.parse('2026-03-19T10:06:00Z'),
        metadata: const <String, Object?>{'project_id': 'project-1'},
      ),
    );
    await Future<void>.delayed(const Duration(milliseconds: 20));

    expect(controller.projects.first.threadCount, 2);
    await streamController.close();
    controller.dispose();
  });
}

class _FakeRestClient implements CodeScopeRestClient {
  _FakeRestClient(this.projects);

  List<ProjectRecord> projects;

  @override
  Future<List<ProjectRecord>> fetchProjects() async => projects;

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
    throw UnimplementedError();
  }

  @override
  Future<List<ThreadMessageRecord>> fetchThreadMessages(String threadId) {
    throw UnimplementedError();
  }

  @override
  Future<List<PromptCommandTask>> fetchThreadCommands(String threadId) {
    throw UnimplementedError();
  }

  @override
  Future<PromptCommandTask> sendThreadPrompt(String threadId, String content) {
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
  Future<List<FileTreeNode>> fetchProjectFileTree(String projectId) {
    throw UnimplementedError();
  }

  @override
  Future<FileContentRecord> fetchSessionFileContent(
    String sessionId,
    String path,
  ) {
    throw UnimplementedError();
  }

  @override
  Future<FileContentRecord> fetchProjectFileContent(
    String projectId,
    String path,
  ) {
    throw UnimplementedError();
  }
}

class _MutableProjectRestClient extends _FakeRestClient {
  _MutableProjectRestClient()
      : super(<ProjectRecord>[
          ProjectRecord(
            id: 'project-1',
            name: 'codeScope',
            workspaceRoot: 'D:/Work/Code/Cross/codeScope',
            machineId: 'devbox',
            threadCount: 1,
            runningThreadCount: 1,
            createdAt: DateTime.parse('2026-03-19T09:45:00Z'),
            lastActivityAt: DateTime.parse('2026-03-19T10:05:00Z'),
          ),
        ]);
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
