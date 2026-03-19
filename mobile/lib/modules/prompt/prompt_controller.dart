import 'dart:async';

import 'package:flutter/foundation.dart';

import '../log/log_event.dart';
import '../../services/rest_client.dart';
import '../../services/websocket_client.dart';
import 'prompt_command_task.dart';

class PromptController extends ChangeNotifier {
  PromptController({
    required this.sessionId,
    required this.restClient,
    required this.webSocketClient,
  });

  final String sessionId;
  final CodeScopeRestClient restClient;
  final CodeScopeWebSocketClient webSocketClient;

  bool _loading = false;
  bool _sending = false;
  String? _errorMessage;
  List<PromptCommandTask> _tasks = const <PromptCommandTask>[];
  List<LogEvent> _events = const <LogEvent>[];
  StreamSubscription<LogEvent>? _subscription;

  bool get isLoading => _loading;
  bool get isSending => _sending;
  String? get errorMessage => _errorMessage;
  List<PromptCommandTask> get tasks => _tasks;
  List<LogEvent> get events => _events;

  Future<void> load() async {
    _loading = true;
    _errorMessage = null;
    notifyListeners();

    await _subscription?.cancel();

    try {
      final tasks = await restClient.fetchSessionCommands(sessionId);
      final events = await restClient.fetchSessionEvents(sessionId);
      _tasks = tasks
        ..sort((PromptCommandTask left, PromptCommandTask right) {
          return right.createdAt.compareTo(left.createdAt);
        });
      _events = events;
      _subscription = webSocketClient.subscribeToSession(sessionId).listen(
        _handleLiveEvent,
        onError: (Object error) {
          _errorMessage = error.toString();
          notifyListeners();
        },
      );
    } catch (error) {
      _errorMessage = error.toString();
    } finally {
      _loading = false;
      notifyListeners();
    }
  }

  Future<void> sendPrompt(String content) async {
    final trimmed = content.trim();
    if (trimmed.isEmpty) {
      return;
    }

    _sending = true;
    _errorMessage = null;
    notifyListeners();

    try {
      final task = await restClient.sendPrompt(sessionId, trimmed);
      _tasks = <PromptCommandTask>[task, ..._tasks];
    } catch (error) {
      _errorMessage = error.toString();
    } finally {
      _sending = false;
      notifyListeners();
    }
  }

  void _handleLiveEvent(LogEvent event) {
    _events = <LogEvent>[..._events, event];
    if (event.commandId != null && event.type == LogEventType.commandResult) {
      _tasks = _tasks.map((PromptCommandTask task) {
        if (task.id != event.commandId) {
          return task;
        }
        return task.copyWith(
          status: event.commandStatus == 'success'
              ? PromptCommandTaskStatus.success
              : PromptCommandTaskStatus.failed,
          result: event.content,
          updatedAt: event.createdAt,
        );
      }).toList(growable: false);
    }
    notifyListeners();
  }

  Future<void> disposeAsync() async {
    await _subscription?.cancel();
    super.dispose();
  }

  @override
  void dispose() {
    _subscription?.cancel();
    super.dispose();
  }
}
