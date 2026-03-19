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
  });

  final String id;
  final String threadId;
  final ThreadMessageRole role;
  final String content;
  final DateTime createdAt;
  final int sequence;
  final String? sourceType;
}
