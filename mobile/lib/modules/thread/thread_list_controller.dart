import 'dart:async';

import 'package:flutter/foundation.dart';

import '../../services/rest_client.dart';
import '../../services/websocket_client.dart';
import 'thread_record.dart';

class ThreadListController extends ChangeNotifier {
  ThreadListController(this._restClient, this._webSocketClient);

  final CodeScopeRestClient _restClient;
  final CodeScopeWebSocketClient _webSocketClient;

  bool _loading = false;
  String? _errorMessage;
  List<ThreadRecord> _threads = const <ThreadRecord>[];
  StreamSubscription? _subscription;
  String? _projectId;

  bool get isLoading => _loading;
  String? get errorMessage => _errorMessage;
  List<ThreadRecord> get threads => _threads;

  Future<void> loadThreads(String projectId) async {
    _projectId = projectId;
    _loading = true;
    _errorMessage = null;
    notifyListeners();

    try {
      final threads = await _restClient.fetchProjectThreads(projectId);
      _threads = List<ThreadRecord>.from(threads)
        ..sort(
          (ThreadRecord left, ThreadRecord right) =>
              right.lastActivityAt.compareTo(left.lastActivityAt),
        );
    } catch (error) {
      _errorMessage = error.toString();
    } finally {
      _loading = false;
      notifyListeners();
    }
  }

  Future<void> startRealtime(String projectId) async {
    _projectId = projectId;
    await _subscription?.cancel();
    _subscription = _webSocketClient.subscribeToProject(projectId).listen(
      (event) {
        if (_projectId != null) {
          loadThreads(_projectId!);
        }
      },
      onError: (Object error) {
        _errorMessage = error.toString();
        notifyListeners();
      },
    );
  }

  @override
  void dispose() {
    _subscription?.cancel();
    super.dispose();
  }
}
