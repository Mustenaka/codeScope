import '../../modules/log/log_event.dart';
import '../../modules/file/file_content_record.dart';
import '../../modules/file/file_tree_node.dart';
import '../../modules/message/thread_message_record.dart';
import '../../modules/project/project_record.dart';
import '../../modules/prompt/prompt_command_task.dart';
import '../../modules/session/session_record.dart';
import '../../modules/thread/thread_record.dart';
import '../rest_client.dart';
import 'mock_data_provider.dart';

class MockCodeScopeRestClient implements CodeScopeRestClient {
  MockCodeScopeRestClient(this._provider);

  final MockDataProvider _provider;

  @override
  Future<List<ProjectRecord>> fetchProjects() async {
    await Future<void>.delayed(const Duration(milliseconds: 250));
    return _provider.buildProjects();
  }

  @override
  Future<ProjectRecord> fetchProjectDetail(String projectId) async {
    await Future<void>.delayed(const Duration(milliseconds: 150));
    return _provider.buildProjectDetail(projectId);
  }

  @override
  Future<List<ThreadRecord>> fetchProjectThreads(String projectId) async {
    await Future<void>.delayed(const Duration(milliseconds: 150));
    return _provider.buildThreads(projectId);
  }

  @override
  Future<ThreadRecord> createProjectThread(String projectId, String content) async {
    await Future<void>.delayed(const Duration(milliseconds: 150));
    return _provider.createProjectThread(projectId, content);
  }

  @override
  Future<ThreadRecord> fetchThreadDetail(String threadId) async {
    await Future<void>.delayed(const Duration(milliseconds: 150));
    return _provider.buildThreadDetail(threadId);
  }

  @override
  Future<List<ThreadMessageRecord>> fetchThreadMessages(String threadId) async {
    await Future<void>.delayed(const Duration(milliseconds: 150));
    return _provider.buildThreadMessages(threadId);
  }

  @override
  Future<List<PromptCommandTask>> fetchThreadCommands(String threadId) async {
    await Future<void>.delayed(const Duration(milliseconds: 150));
    return _provider.buildPromptCommands(threadId);
  }

  @override
  Future<PromptCommandTask> sendThreadPrompt(
      String threadId, String content) async {
    await Future<void>.delayed(const Duration(milliseconds: 150));
    return _provider.createPromptCommand(threadId, content);
  }

  @override
  Future<List<SessionRecord>> fetchSessions() async {
    await Future<void>.delayed(const Duration(milliseconds: 250));
    return _provider.buildSessions();
  }

  @override
  Future<SessionRecord> fetchSessionDetail(String sessionId) async {
    await Future<void>.delayed(const Duration(milliseconds: 150));
    return _provider.buildSessionDetail(sessionId);
  }

  @override
  Future<List<LogEvent>> fetchSessionEvents(String sessionId) async {
    await Future<void>.delayed(const Duration(milliseconds: 150));
    return _provider.buildInitialEvents(sessionId);
  }

  @override
  Future<List<PromptCommandTask>> fetchSessionCommands(String sessionId) async {
    await Future<void>.delayed(const Duration(milliseconds: 150));
    return _provider.buildPromptCommands(sessionId);
  }

  @override
  Future<PromptCommandTask> sendPrompt(String sessionId, String content) async {
    await Future<void>.delayed(const Duration(milliseconds: 150));
    return _provider.createPromptCommand(sessionId, content);
  }

  @override
  Future<List<FileTreeNode>> fetchSessionFileTree(String sessionId) async {
    await Future<void>.delayed(const Duration(milliseconds: 150));
    return _provider.buildFileTree(sessionId);
  }

  @override
  Future<List<FileTreeNode>> fetchProjectFileTree(String projectId) async {
    await Future<void>.delayed(const Duration(milliseconds: 150));
    return _provider.buildFileTree(projectId);
  }

  @override
  Future<FileContentRecord> fetchSessionFileContent(
    String sessionId,
    String path,
  ) async {
    await Future<void>.delayed(const Duration(milliseconds: 120));
    return _provider.buildFileContent(sessionId, path);
  }

  @override
  Future<FileContentRecord> fetchProjectFileContent(
    String projectId,
    String path,
  ) async {
    await Future<void>.delayed(const Duration(milliseconds: 120));
    return _provider.buildFileContent(projectId, path);
  }
}
