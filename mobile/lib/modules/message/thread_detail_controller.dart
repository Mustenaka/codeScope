import 'dart:async';

import 'package:flutter/foundation.dart';

import '../../services/rest_client.dart';
import '../../services/websocket_client.dart';
import '../thread/thread_record.dart';
import 'thread_message_record.dart';

class ThreadDetailController extends ChangeNotifier {
  ThreadDetailController(this._restClient, this._webSocketClient);

  final CodeScopeRestClient _restClient;
  final CodeScopeWebSocketClient _webSocketClient;

  bool _loading = false;
  String? _errorMessage;
  ThreadRecord? _thread;
  List<ThreadMessageRecord> _messages = const <ThreadMessageRecord>[];
  StreamSubscription<ThreadMessageRecord>? _subscription;

  bool get isLoading => _loading;
  String? get errorMessage => _errorMessage;
  ThreadRecord? get thread => _thread;
  List<ThreadMessageRecord> get messages => _messages;

  Future<void> load(String threadId) async {
    _loading = true;
    _errorMessage = null;
    notifyListeners();
    await _subscription?.cancel();

    try {
      _thread = await _restClient.fetchThreadDetail(threadId);
      _messages = await _restClient.fetchThreadMessages(threadId);
      _subscription = _webSocketClient.subscribeToThread(threadId).listen(
        _appendMessage,
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

  void _appendMessage(ThreadMessageRecord message) {
    _messages = List<ThreadMessageRecord>.from(_messages)..add(message);
    _messages.sort(
      (ThreadMessageRecord left, ThreadMessageRecord right) =>
          left.createdAt.compareTo(right.createdAt),
    );
    notifyListeners();
  }

  @override
  void dispose() {
    _subscription?.cancel();
    super.dispose();
  }
}
