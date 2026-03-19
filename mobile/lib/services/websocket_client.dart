import '../modules/log/log_event.dart';
import '../modules/message/thread_message_record.dart';

abstract class CodeScopeWebSocketClient {
  Stream<LogEvent> subscribeToProjects();

  Stream<LogEvent> subscribeToProject(String projectId);

  Stream<LogEvent> subscribeToSession(String sessionId);

  Stream<ThreadMessageRecord> subscribeToThread(String threadId);
}

class UnimplementedCodeScopeWebSocketClient
    implements CodeScopeWebSocketClient {
  const UnimplementedCodeScopeWebSocketClient(this.url);

  final String url;

  @override
  Stream<LogEvent> subscribeToProjects() {
    throw UnimplementedError(
      'Connect subscribeToProjects to WebSocket $url.',
    );
  }

  @override
  Stream<LogEvent> subscribeToProject(String projectId) {
    throw UnimplementedError(
      'Connect subscribeToProject to WebSocket $url for project $projectId.',
    );
  }

  @override
  Stream<LogEvent> subscribeToSession(String sessionId) {
    throw UnimplementedError(
      'Connect subscribeToSession to WebSocket $url for session $sessionId.',
    );
  }

  @override
  Stream<ThreadMessageRecord> subscribeToThread(String threadId) {
    throw UnimplementedError(
      'Connect subscribeToThread to WebSocket $url for thread $threadId.',
    );
  }
}
