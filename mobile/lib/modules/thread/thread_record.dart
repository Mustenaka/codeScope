enum ThreadStatus {
  running,
  waitingPrompt,
  waitingReview,
  completed,
  blocked,
  offline,
  stale;

  String get label {
    switch (this) {
      case ThreadStatus.running:
        return 'Running';
      case ThreadStatus.waitingPrompt:
        return 'Waiting Prompt';
      case ThreadStatus.waitingReview:
        return 'Waiting Review';
      case ThreadStatus.completed:
        return 'Completed';
      case ThreadStatus.blocked:
        return 'Blocked';
      case ThreadStatus.offline:
        return 'Offline';
      case ThreadStatus.stale:
        return 'Stale';
    }
  }
}

class ThreadRecord {
  const ThreadRecord({
    required this.id,
    required this.projectId,
    required this.sessionId,
    required this.title,
    required this.status,
    required this.lastActivityAt,
    required this.startedAt,
    this.agentKind,
    this.summary = '',
    this.endedAt,
  });

  final String id;
  final String projectId;
  final String sessionId;
  final String title;
  final String? agentKind;
  final ThreadStatus status;
  final String summary;
  final DateTime lastActivityAt;
  final DateTime startedAt;
  final DateTime? endedAt;

  String get agentLabel {
    final normalized = agentKind?.trim().toLowerCase();
    switch (normalized) {
      case 'codex':
        return 'Codex';
      case 'claude':
        return 'Claude';
      case '':
      case null:
        return 'Agent';
      default:
        return agentKind!;
    }
  }

  bool get hasSummary => summary.trim().isNotEmpty;

  String get displaySummary {
    if (hasSummary) {
      return summary;
    }

    switch (status) {
      case ThreadStatus.running:
        return 'Running. No readable message summary has been captured yet.';
      case ThreadStatus.waitingPrompt:
        return 'Waiting for the next prompt. No readable summary has been captured yet.';
      case ThreadStatus.waitingReview:
        return 'Waiting for review. No readable summary has been captured yet.';
      case ThreadStatus.completed:
        return 'Completed. No readable summary was captured for this thread.';
      case ThreadStatus.blocked:
        return 'Blocked. No readable summary has been captured yet.';
      case ThreadStatus.offline:
        return 'Offline. The last known thread state is preserved until the bridge reconnects.';
      case ThreadStatus.stale:
        return 'Stale. The thread has gone quiet longer than the active window.';
    }
  }
}
