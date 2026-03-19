enum ThreadMessageRole {
  user,
  assistant,
  system;

  String get label {
    switch (this) {
      case ThreadMessageRole.user:
        return 'You';
      case ThreadMessageRole.assistant:
        return 'Agent';
      case ThreadMessageRole.system:
        return 'System';
    }
  }
}

class ThreadMessageRecord {
  const ThreadMessageRecord({
    required this.id,
    required this.threadId,
    required this.role,
    required this.content,
    required this.createdAt,
    required this.sequence,
    this.sourceType,
    this.agentKind,
  });

  final String id;
  final String threadId;
  final ThreadMessageRole role;
  final String content;
  final DateTime createdAt;
  final int sequence;
  final String? sourceType;
  final String? agentKind;

  String get sourceLabel {
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
}
