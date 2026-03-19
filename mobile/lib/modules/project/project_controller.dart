import 'dart:async';

import 'package:flutter/foundation.dart';

import '../../services/rest_client.dart';
import '../../services/websocket_client.dart';
import 'project_record.dart';

class ProjectController extends ChangeNotifier {
  ProjectController(this._restClient, this._webSocketClient);

  final CodeScopeRestClient _restClient;
  final CodeScopeWebSocketClient _webSocketClient;

  bool _loading = false;
  String? _errorMessage;
  List<ProjectRecord> _projects = const <ProjectRecord>[];
  StreamSubscription? _subscription;

  bool get isLoading => _loading;
  String? get errorMessage => _errorMessage;
  List<ProjectRecord> get projects => _projects;

  Future<void> loadProjects() async {
    _loading = true;
    _errorMessage = null;
    notifyListeners();

    try {
      final projects = await _restClient.fetchProjects();
      _projects = List<ProjectRecord>.from(projects)
        ..sort(
          (ProjectRecord left, ProjectRecord right) =>
              right.lastActivityAt.compareTo(left.lastActivityAt),
        );
    } catch (error) {
      _errorMessage = error.toString();
    } finally {
      _loading = false;
      notifyListeners();
    }
  }

  Future<void> startRealtime() async {
    await _subscription?.cancel();
    _subscription = _webSocketClient.subscribeToProjects().listen(
      (event) {
        loadProjects();
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
