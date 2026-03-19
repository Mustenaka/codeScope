import 'dart:async';

import 'package:flutter/foundation.dart';

import '../../services/rest_client.dart';
import '../../services/websocket_client.dart';
import '../session/session_record.dart';
import '../session/session_signal_analyzer.dart';
import 'log_event.dart';

class LogController extends ChangeNotifier {
  LogController({
    required this.restClient,
    required this.webSocketClient,
  });

  final CodeScopeRestClient restClient;
  final CodeScopeWebSocketClient webSocketClient;

  bool _loading = false;
  String? _errorMessage;
  SessionRecord? _session;
  List<LogEvent> _events = const <LogEvent>[];
  StreamSubscription<LogEvent>? _subscription;

  bool get isLoading => _loading;
  String? get errorMessage => _errorMessage;
  SessionRecord? get session => _session;
  List<LogEvent> get events => _events;
  SessionDigest? get digest {
    final session = _session;
    if (session == null) {
      return null;
    }
    return SessionSignalAnalyzer.analyze(session, _events);
  }

  Future<void> load(String sessionId) async {
    _loading = true;
    _errorMessage = null;
    notifyListeners();

    await _subscription?.cancel();

    try {
      final session = await restClient.fetchSessionDetail(sessionId);
      final events = await restClient.fetchSessionEvents(sessionId);

      _session = session;
      _events = events..sort((a, b) => a.createdAt.compareTo(b.createdAt));

      _subscription = webSocketClient.subscribeToSession(sessionId).listen(
        _appendEvent,
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

  void _appendEvent(LogEvent event) {
    _events = List<LogEvent>.from(_events)..add(event);
    notifyListeners();
  }

  @override
  void dispose() {
    _subscription?.cancel();
    super.dispose();
  }
}
