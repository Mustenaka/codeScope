enum ThreadStatus {
  running,
  waitingPrompt,
  waitingReview,
  completed,
  blocked;

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
    }
  }
}
