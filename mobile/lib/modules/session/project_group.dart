import 'session_record.dart';

class ProjectGroup {
  const ProjectGroup({
    required this.projectName,
    required this.workspaceRoot,
    required this.sessions,
  });

  final String projectName;
  final String workspaceRoot;
  final List<SessionRecord> sessions;

  DateTime get updatedAt => sessions
      .map((SessionRecord session) => session.updatedAt)
      .reduce((DateTime left, DateTime right) =>
          left.isAfter(right) ? left : right);

  int get runningCount => sessions
      .where(
        (SessionRecord session) =>
            session.status == SessionStatus.running &&
            session.agentState == AgentSummaryState.running,
      )
      .length;

  int get waitingPromptCount => sessions
      .where(
        (SessionRecord session) =>
            session.status == SessionStatus.running &&
            session.agentState == AgentSummaryState.waitingPrompt,
      )
      .length;

  int get waitingReviewCount => sessions
      .where(
        (SessionRecord session) =>
            session.status == SessionStatus.running &&
            session.agentState == AgentSummaryState.waitingReview,
      )
      .length;
}
