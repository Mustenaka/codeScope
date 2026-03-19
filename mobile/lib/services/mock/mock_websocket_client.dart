import '../../modules/log/log_event.dart';
import '../../modules/message/thread_message_record.dart';
import '../websocket_client.dart';
import 'mock_data_provider.dart';

class MockCodeScopeWebSocketClient implements CodeScopeWebSocketClient {
  MockCodeScopeWebSocketClient(this._provider);

  final MockDataProvider _provider;

  @override
  Stream<LogEvent> subscribeToProjects() => const Stream.empty();

  @override
  Stream<LogEvent> subscribeToProject(String projectId) => const Stream.empty();

  @override
  Stream<LogEvent> subscribeToSession(String sessionId) {
    return _provider.buildLiveEvents(sessionId);
  }

  @override
  Stream<ThreadMessageRecord> subscribeToThread(String threadId) async* {
    for (final message in _provider.buildThreadMessages(threadId)) {
      yield message;
    }
  }
}
