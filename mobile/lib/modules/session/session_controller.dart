import 'package:flutter/foundation.dart';

import '../log/log_event.dart';
import '../../services/rest_client.dart';
import 'project_group.dart';
import 'session_record.dart';
import 'session_signal_analyzer.dart';

class SessionController extends ChangeNotifier {
  SessionController(this._restClient);

  final CodeScopeRestClient _restClient;

  bool _loading = false;
  String? _errorMessage;
  List<SessionRecord> _sessions = const <SessionRecord>[];

  bool get isLoading => _loading;
  String? get errorMessage => _errorMessage;
  List<SessionRecord> get sessions => _sessions;
  List<ProjectGroup> get projectGroups {
    final grouped = <String, List<SessionRecord>>{};
    for (final SessionRecord session in _sessions) {
      grouped
          .putIfAbsent(session.workspaceRoot, () => <SessionRecord>[])
          .add(session);
    }

    final groups = grouped.entries
        .map((MapEntry<String, List<SessionRecord>> entry) {
      final sessions = List<SessionRecord>.from(entry.value)
        ..sort((SessionRecord left, SessionRecord right) =>
            right.updatedAt.compareTo(left.updatedAt));
      return ProjectGroup(
        projectName: sessions.first.projectName,
        workspaceRoot: entry.key,
        sessions: sessions,
      );
    }).toList()
      ..sort((ProjectGroup left, ProjectGroup right) =>
          right.updatedAt.compareTo(left.updatedAt));
    return groups;
  }

  Future<void> loadSessions() async {
    _loading = true;
    _errorMessage = null;
    notifyListeners();

    try {
      final sessions = await _restClient.fetchSessions();
      final enriched = <SessionRecord>[];
      for (final SessionRecord session in sessions) {
        List<LogEvent> events;
        try {
          events = await _restClient.fetchSessionEvents(session.id);
        } catch (_) {
          events = const <LogEvent>[];
        }
        final digest = SessionSignalAnalyzer.analyze(session, events);
        enriched.add(
          session.copyWith(
            summary: digest.summary,
            agentState: switch (digest.state) {
              AgentUserState.running => AgentSummaryState.running,
              AgentUserState.waitingPrompt => AgentSummaryState.waitingPrompt,
              AgentUserState.waitingReview => AgentSummaryState.waitingReview,
              AgentUserState.completed => AgentSummaryState.completed,
              AgentUserState.blocked => AgentSummaryState.blocked,
            },
          ),
        );
      }
      _sessions = enriched;
    } catch (error) {
      _errorMessage = error.toString();
    } finally {
      _loading = false;
      notifyListeners();
    }
  }
}
