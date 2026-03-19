import 'package:codescope_mobile/modules/file/file_browser_controller.dart';
import 'package:codescope_mobile/modules/file/file_content_record.dart';
import 'package:codescope_mobile/modules/file/file_tree_node.dart';
import 'package:codescope_mobile/modules/log/log_event.dart';
import 'package:codescope_mobile/modules/message/thread_message_record.dart';
import 'package:codescope_mobile/modules/project/project_record.dart';
import 'package:codescope_mobile/modules/prompt/prompt_command_task.dart';
import 'package:codescope_mobile/modules/session/session_record.dart';
import 'package:codescope_mobile/modules/thread/thread_record.dart';
import 'package:codescope_mobile/services/rest_client.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('loads file tree and file preview from rest client', () async {
    final controller = FileBrowserController(
      restClient: _FakeRestClient(),
      projectId: 'project-001',
    );

    await controller.load();
    await controller.selectFile('lib/main.dart');

    expect(controller.tree, hasLength(1));
    expect(controller.tree.single.children.single.path, 'lib/main.dart');
    expect(controller.selectedContent?.content, 'void main() {}');
  });
}

class _FakeRestClient implements CodeScopeRestClient {
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
  Future<List<FileTreeNode>> fetchSessionFileTree(String sessionId) async {
    throw UnimplementedError();
  }

  @override
  Future<List<FileTreeNode>> fetchProjectFileTree(String projectId) async {
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
    ];
  }

  @override
  Future<FileContentRecord> fetchSessionFileContent(
    String sessionId,
    String path,
  ) async {
    throw UnimplementedError();
  }

  @override
  Future<FileContentRecord> fetchProjectFileContent(
    String projectId,
    String path,
  ) async {
    return const FileContentRecord(
      path: 'lib/main.dart',
      size: 128,
      previewable: true,
      language: 'dart',
      content: 'void main() {}',
    );
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
