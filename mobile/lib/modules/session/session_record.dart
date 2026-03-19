enum SessionStatus {
  created,
  running,
  stopped,
  failed;

  String get label {
    switch (this) {
      case SessionStatus.created:
        return 'Created';
      case SessionStatus.running:
        return 'Running';
      case SessionStatus.stopped:
        return 'Stopped';
      case SessionStatus.failed:
        return 'Failed';
    }
  }
}

enum AgentSummaryState {
  running,
  waitingPrompt,
  waitingReview,
  completed,
  blocked;
}

class SessionRecord {
  const SessionRecord({
    required this.id,
    required this.projectName,
    required this.workspaceRoot,
    required this.machineId,
    required this.status,
    required this.startedAt,
    required this.updatedAt,
    this.endedAt,
    this.summary = '',
    this.agentState = AgentSummaryState.running,
  });

  final String id;
  final String projectName;
  final String workspaceRoot;
  final String machineId;
  final SessionStatus status;
  final DateTime startedAt;
  final DateTime updatedAt;
  final DateTime? endedAt;
  final String summary;
  final AgentSummaryState agentState;

  SessionRecord copyWith({
    String? summary,
    AgentSummaryState? agentState,
  }) {
    return SessionRecord(
      id: id,
      projectName: projectName,
      workspaceRoot: workspaceRoot,
      machineId: machineId,
      status: status,
      startedAt: startedAt,
      updatedAt: updatedAt,
      endedAt: endedAt,
      summary: summary ?? this.summary,
      agentState: agentState ?? this.agentState,
    );
  }
}
