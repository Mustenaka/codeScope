import '../modules/log/log_event.dart';
import '../modules/file/file_content_record.dart';
import '../modules/file/file_tree_node.dart';
import '../modules/message/thread_message_record.dart';
import '../modules/project/project_record.dart';
import '../modules/prompt/prompt_command_task.dart';
import '../modules/session/session_record.dart';
import '../modules/thread/thread_record.dart';

abstract class CodeScopeRestClient {
  Future<List<ProjectRecord>> fetchProjects();

  Future<ProjectRecord> fetchProjectDetail(String projectId);

  Future<List<ThreadRecord>> fetchProjectThreads(String projectId);

  Future<ThreadRecord> fetchThreadDetail(String threadId);

  Future<List<ThreadMessageRecord>> fetchThreadMessages(String threadId);

  Future<List<SessionRecord>> fetchSessions();

  Future<SessionRecord> fetchSessionDetail(String sessionId);

  Future<List<LogEvent>> fetchSessionEvents(String sessionId);

  Future<List<PromptCommandTask>> fetchSessionCommands(String sessionId);

  Future<PromptCommandTask> sendPrompt(String sessionId, String content);

  Future<List<FileTreeNode>> fetchSessionFileTree(String sessionId);

  Future<FileContentRecord> fetchSessionFileContent(
      String sessionId, String path);
}

class UnimplementedCodeScopeRestClient implements CodeScopeRestClient {
  const UnimplementedCodeScopeRestClient(this.baseUrl);

  final String baseUrl;

  @override
  Future<List<ProjectRecord>> fetchProjects() {
    throw UnimplementedError(
      'Connect fetchProjects to GET $baseUrl/projects.',
    );
  }

  @override
  Future<ProjectRecord> fetchProjectDetail(String projectId) {
    throw UnimplementedError(
      'Connect fetchProjectDetail to GET $baseUrl/projects/$projectId.',
    );
  }

  @override
  Future<List<ThreadRecord>> fetchProjectThreads(String projectId) {
    throw UnimplementedError(
      'Connect fetchProjectThreads to GET $baseUrl/projects/$projectId/threads.',
    );
  }

  @override
  Future<ThreadRecord> fetchThreadDetail(String threadId) {
    throw UnimplementedError(
      'Connect fetchThreadDetail to GET $baseUrl/threads/$threadId.',
    );
  }

  @override
  Future<List<ThreadMessageRecord>> fetchThreadMessages(String threadId) {
    throw UnimplementedError(
      'Connect fetchThreadMessages to GET $baseUrl/threads/$threadId/messages.',
    );
  }

  @override
  Future<List<SessionRecord>> fetchSessions() {
    throw UnimplementedError(
      'Connect fetchSessions to GET $baseUrl/sessions.',
    );
  }

  @override
  Future<SessionRecord> fetchSessionDetail(String sessionId) {
    throw UnimplementedError(
      'Connect fetchSessionDetail to GET $baseUrl/sessions/$sessionId.',
    );
  }

  @override
  Future<List<LogEvent>> fetchSessionEvents(String sessionId) {
    throw UnimplementedError(
      'Connect fetchSessionEvents to GET $baseUrl/sessions/$sessionId/events.',
    );
  }

  @override
  Future<List<PromptCommandTask>> fetchSessionCommands(String sessionId) {
    throw UnimplementedError(
      'Connect fetchSessionCommands to GET $baseUrl/sessions/$sessionId/commands.',
    );
  }

  @override
  Future<PromptCommandTask> sendPrompt(String sessionId, String content) {
    throw UnimplementedError(
      'Connect sendPrompt to POST $baseUrl/sessions/$sessionId/commands/prompt.',
    );
  }

  @override
  Future<List<FileTreeNode>> fetchSessionFileTree(String sessionId) {
    throw UnimplementedError(
      'Connect fetchSessionFileTree to GET $baseUrl/sessions/$sessionId/files/tree.',
    );
  }

  @override
  Future<FileContentRecord> fetchSessionFileContent(
      String sessionId, String path) {
    throw UnimplementedError(
      'Connect fetchSessionFileContent to GET $baseUrl/sessions/$sessionId/files/content?path=$path.',
    );
  }
}
