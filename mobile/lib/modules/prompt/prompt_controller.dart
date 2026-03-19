import 'dart:async';

import 'package:flutter/foundation.dart';

import '../log/log_event.dart';
import '../thread/thread_record.dart';
import '../../services/rest_client.dart';
import '../../services/real/server_error_presenter.dart';
import '../../services/websocket_client.dart';
import 'prompt_command_task.dart';

class PromptController extends ChangeNotifier {
  PromptController({
    this.sessionId,
    this.threadId,
    required this.restClient,
    required this.webSocketClient,
  }) : assert(
          (sessionId != null && sessionId != '') ||
              (threadId != null && threadId != ''),
          'PromptController requires either sessionId or threadId.',
        );

  final String? sessionId;
  final String? threadId;
  final CodeScopeRestClient restClient;
  final CodeScopeWebSocketClient webSocketClient;

  bool _loading = false;
  bool _sending = false;
  String? _errorMessage;
  List<PromptCommandTask> _tasks = const <PromptCommandTask>[];
  List<LogEvent> _events = const <LogEvent>[];
  ThreadRecord? _thread;
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
      final resolvedSessionID = await _resolveSessionID();
      final tasks = threadId != null
          ? await restClient.fetchThreadCommands(threadId!)
          : await restClient.fetchSessionCommands(resolvedSessionID);
      final events = await restClient.fetchSessionEvents(resolvedSessionID);
      _tasks = tasks
        ..sort((PromptCommandTask left, PromptCommandTask right) {
          return right.createdAt.compareTo(left.createdAt);
        });
      _events = events;
      _subscription = webSocketClient.subscribeToSession(resolvedSessionID).listen(
        _handleLiveEvent,
        onError: (Object error) {
          _errorMessage = error.toString();
          notifyListeners();
        },
      );
    } catch (error) {
      _errorMessage = presentServerError(error);
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
      final task = threadId != null
          ? await restClient.sendThreadPrompt(threadId!, trimmed)
          : await restClient.sendPrompt(sessionId!, trimmed);
      _tasks = <PromptCommandTask>[task, ..._tasks];
    } catch (error) {
      _errorMessage = presentServerError(error);
    } finally {
      _sending = false;
      notifyListeners();
    }
  }

  Future<String> _resolveSessionID() async {
    if (sessionId != null && sessionId!.isNotEmpty) {
      return sessionId!;
    }
    if (_thread != null) {
      return _thread!.sessionId;
    }
    final detail = await restClient.fetchThreadDetail(threadId!);
    _thread = detail;
    return detail.sessionId;
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
