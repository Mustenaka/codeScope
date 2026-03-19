import 'package:codescope_mobile/app/app_environment.dart';
import 'package:codescope_mobile/app/app_scope.dart';
import 'package:codescope_mobile/app/app_settings.dart';
import 'package:codescope_mobile/modules/file/file_content_record.dart';
import 'package:codescope_mobile/modules/file/file_tree_node.dart';
import 'package:codescope_mobile/modules/log/log_event.dart';
import 'package:codescope_mobile/modules/message/thread_detail_page.dart';
import 'package:codescope_mobile/modules/message/thread_message_record.dart';
import 'package:codescope_mobile/modules/project/project_record.dart';
import 'package:codescope_mobile/modules/prompt/prompt_command_task.dart';
import 'package:codescope_mobile/modules/session/session_record.dart';
import 'package:codescope_mobile/modules/thread/thread_record.dart';
import 'package:codescope_mobile/services/app_services.dart';
import 'package:codescope_mobile/services/connection_tester.dart';
import 'package:codescope_mobile/services/rest_client.dart';
import 'package:codescope_mobile/services/websocket_client.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  testWidgets('opens a full message sheet for long assistant content', (
    WidgetTester tester,
  ) async {
    final thread = ThreadRecord(
      id: 'thread-001',
      projectId: 'project-001',
      sessionId: 'session-001',
      title: 'Long thread',
      agentKind: 'claude',
      status: ThreadStatus.waitingReview,
      summary: 'A very long summary',
      lastActivityAt: DateTime.parse('2026-03-19T10:05:00Z'),
      startedAt: DateTime.parse('2026-03-19T10:00:00Z'),
    );

    await tester.pumpWidget(
      _buildHarness(
        restClient: _ThreadPageRestClient(thread),
        webSocketClient: _ThreadPageSocketClient(),
        child: ThreadDetailPage(thread: thread),
      ),
    );

    await tester.pumpAndSettle();

    expect(find.text('View Full Message'), findsOneWidget);

    await tester.tap(find.text('View Full Message'));
    await tester.pumpAndSettle();

    expect(find.text('Full Message'), findsOneWidget);
    expect(find.textContaining('Line 30'), findsWidgets);
  });
}

Widget _buildHarness({
  required CodeScopeRestClient restClient,
  required CodeScopeWebSocketClient webSocketClient,
  required Widget child,
}) {
  final services = AppServices(
    environment: AppEnvironment.mock,
    restClient: restClient,
    webSocketClient: webSocketClient,
    connectionTester: _NoopConnectionTester(),
  );
  return AppScope(
    services: services,
    settingsController: AppSettingsController(AppEnvironment.mock),
    child: MaterialApp(home: child),
  );
}

class _ThreadPageRestClient implements CodeScopeRestClient {
  _ThreadPageRestClient(this.thread);

  final ThreadRecord thread;

  @override
  Future<ThreadRecord> fetchThreadDetail(String threadId) async => thread;

  @override
  Future<List<ThreadMessageRecord>> fetchThreadMessages(String threadId) async {
    final content = List<String>.generate(30, (int index) => 'Line ${index + 1}')
        .join('\n');
    return <ThreadMessageRecord>[
      ThreadMessageRecord(
        id: 'message-001',
        threadId: threadId,
        role: ThreadMessageRole.assistant,
        content: content,
        createdAt: DateTime.parse('2026-03-19T10:05:00Z'),
        sequence: 1,
        sourceType: 'ai_output',
        agentKind: 'claude',
      ),
    ];
  }

  @override
  Future<List<ProjectRecord>> fetchProjects() => throw UnimplementedError();
  @override
  Future<ProjectRecord> fetchProjectDetail(String projectId) => throw UnimplementedError();
  @override
  Future<List<ThreadRecord>> fetchProjectThreads(String projectId) => throw UnimplementedError();
  @override
  Future<ThreadRecord> createProjectThread(String projectId, String content) => throw UnimplementedError();
  @override
  Future<List<PromptCommandTask>> fetchThreadCommands(String threadId) => throw UnimplementedError();
  @override
  Future<PromptCommandTask> sendThreadPrompt(String threadId, String content) => throw UnimplementedError();
  @override
  Future<List<FileTreeNode>> fetchProjectFileTree(String projectId) => throw UnimplementedError();
  @override
  Future<FileContentRecord> fetchProjectFileContent(String projectId, String path) => throw UnimplementedError();
  @override
  Future<List<SessionRecord>> fetchSessions() => throw UnimplementedError();
  @override
  Future<SessionRecord> fetchSessionDetail(String sessionId) => throw UnimplementedError();
  @override
  Future<List<LogEvent>> fetchSessionEvents(String sessionId) => throw UnimplementedError();
  @override
  Future<List<PromptCommandTask>> fetchSessionCommands(String sessionId) => throw UnimplementedError();
  @override
  Future<PromptCommandTask> sendPrompt(String sessionId, String content) => throw UnimplementedError();
  @override
  Future<List<FileTreeNode>> fetchSessionFileTree(String sessionId) => throw UnimplementedError();
  @override
  Future<FileContentRecord> fetchSessionFileContent(String sessionId, String path) => throw UnimplementedError();
}

class _ThreadPageSocketClient implements CodeScopeWebSocketClient {
  @override
  Stream<LogEvent> subscribeToProject(String projectId) => const Stream<LogEvent>.empty();

  @override
  Stream<LogEvent> subscribeToProjects() => const Stream<LogEvent>.empty();

  @override
  Stream<LogEvent> subscribeToSession(String sessionId) => const Stream<LogEvent>.empty();

  @override
  Stream<ThreadMessageRecord> subscribeToThread(String threadId) =>
      const Stream<ThreadMessageRecord>.empty();
}

class _NoopConnectionTester implements CodeScopeConnectionTester {
  @override
  Future<ConnectionTestResult> test(AppEnvironment environment) async {
    return const ConnectionTestResult(
      status: ConnectionTestStatus.success,
      message: 'ok',
    );
  }
}
