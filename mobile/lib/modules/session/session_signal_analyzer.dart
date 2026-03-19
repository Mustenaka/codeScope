import '../log/log_event.dart';
import 'session_record.dart';

enum AgentUserState {
  running,
  waitingPrompt,
  waitingReview,
  completed,
  blocked,
}

class SessionDigest {
  const SessionDigest({
    required this.state,
    required this.summary,
    required this.visibleEvents,
  });

  final AgentUserState state;
  final String summary;
  final List<LogEvent> visibleEvents;
}

class SessionSignalAnalyzer {
  const SessionSignalAnalyzer._();

  static SessionDigest analyze(SessionRecord session, List<LogEvent> events) {
    final visibleEvents = events.where(_isVisibleEvent).toList(growable: false)
      ..sort((LogEvent left, LogEvent right) =>
          left.createdAt.compareTo(right.createdAt));
    final latest = visibleEvents.isEmpty ? null : visibleEvents.last;
    final summary =
        latest?.content.isNotEmpty == true ? latest!.content : session.summary;

    return SessionDigest(
      state: _deriveState(session, latest),
      summary: summary,
      visibleEvents: visibleEvents,
    );
  }

  static bool _isVisibleEvent(LogEvent event) {
    if (event.type == LogEventType.heartbeat) {
      return false;
    }
    if (event.type == LogEventType.terminalOutput &&
        event.content.startsWith('[bridge] observing')) {
      return false;
    }
    return true;
  }

  static AgentUserState _deriveState(SessionRecord session, LogEvent? latest) {
    if (session.status == SessionStatus.failed) {
      return AgentUserState.blocked;
    }
    if (session.status == SessionStatus.stopped) {
      return AgentUserState.completed;
    }

    final text = '${latest?.content ?? ''} ${latest?.metadata.values.join(' ')}'
        .toLowerCase();
    if (_containsAny(text, const <String>[
      'waiting for next prompt',
      'waiting for prompt',
      'awaiting prompt',
      'next prompt',
      '等待用户prompt',
      '等待prompt',
    ])) {
      return AgentUserState.waitingPrompt;
    }
    if (_containsAny(text, const <String>[
      'confirm',
      'review',
      'check before merge',
      'please check',
      'please review',
      '等待确认',
      '等待检查',
      '请确认',
      '请检查',
    ])) {
      return AgentUserState.waitingReview;
    }
    return AgentUserState.running;
  }

  static bool _containsAny(String text, List<String> patterns) {
    for (final String pattern in patterns) {
      if (text.contains(pattern)) {
        return true;
      }
    }
    return false;
  }
}
