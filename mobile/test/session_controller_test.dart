import 'package:codescope_mobile/modules/session/session_controller.dart';
import 'package:codescope_mobile/modules/session/session_record.dart';
import 'package:codescope_mobile/modules/log/log_event.dart';
import 'package:codescope_mobile/modules/file/file_content_record.dart';
import 'package:codescope_mobile/modules/file/file_tree_node.dart';
import 'package:codescope_mobile/services/rest_client.dart';
import 'package:codescope_mobile/modules/message/thread_message_record.dart';
import 'package:codescope_mobile/modules/project/project_record.dart';
import 'package:codescope_mobile/modules/prompt/prompt_command_task.dart';
import 'package:codescope_mobile/modules/thread/thread_record.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('groups sessions by workspace-root-backed project key', () async {
    final controller = SessionController(_FakeRestClient(<SessionRecord>[
      SessionRecord(
        id: 'session-001',
        projectName: 'codeScope',
        workspaceRoot: 'D:/Work/Code/Cross/codeScope',
        machineId: 'devbox',
        status: SessionStatus.running,
        startedAt: DateTime.parse('2026-03-17T08:00:00Z'),
        updatedAt: DateTime.parse('2026-03-17T08:05:00Z'),
      ),
      SessionRecord(
        id: 'session-002',
        projectName: 'codeScope',
        workspaceRoot: 'D:/Work/Code/Cross/codeScope',
        machineId: 'devbox',
        status: SessionStatus.created,
        startedAt: DateTime.parse('2026-03-17T08:06:00Z'),
        updatedAt: DateTime.parse('2026-03-17T08:07:00Z'),
      ),
      SessionRecord(
        id: 'session-003',
        projectName: 'other',
        workspaceRoot: 'D:/Work/Code/Cross/other',
        machineId: 'devbox',
        status: SessionStatus.failed,
        startedAt: DateTime.parse('2026-03-17T07:00:00Z'),
        updatedAt: DateTime.parse('2026-03-17T07:05:00Z'),
      ),
    ]));

    await controller.loadSessions();

    expect(controller.projectGroups, hasLength(2));
    expect(controller.projectGroups.first.workspaceRoot,
        'D:/Work/Code/Cross/codeScope');
    expect(controller.projectGroups.first.sessions, hasLength(2));
    expect(controller.projectGroups.first.runningCount, 1);
  });
}

class _FakeRestClient implements CodeScopeRestClient {
  _FakeRestClient(this.sessions);

  final List<SessionRecord> sessions;

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
  Future<ThreadRecord> fetchThreadDetail(String threadId) {
    throw UnimplementedError();
  }

  @override
  Future<List<ThreadMessageRecord>> fetchThreadMessages(String threadId) {
    throw UnimplementedError();
  }

  @override
  Future<List<SessionRecord>> fetchSessions() async => sessions;

  @override
  Future<SessionRecord> fetchSessionDetail(String sessionId) {
    throw UnimplementedError();
  }

  @override
  Future<List<LogEvent>> fetchSessionEvents(String sessionId) {
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
  Future<List<FileTreeNode>> fetchSessionFileTree(String sessionId) {
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
}
