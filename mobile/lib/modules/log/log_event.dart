enum LogEventType {
  aiOutput,
  terminalOutput,
  command,
  commandResult,
  fileChange,
  diff,
  error,
  heartbeat;

  String get label {
    switch (this) {
      case LogEventType.aiOutput:
        return 'AI';
      case LogEventType.terminalOutput:
        return 'Terminal';
      case LogEventType.command:
        return 'Command';
      case LogEventType.commandResult:
        return 'Result';
      case LogEventType.fileChange:
        return 'File';
      case LogEventType.diff:
        return 'Diff';
      case LogEventType.error:
        return 'Error';
      case LogEventType.heartbeat:
        return 'Heartbeat';
    }
  }
}

enum LogLevel {
  debug,
  info,
  warning,
  error;
}

class LogEvent {
  const LogEvent({
    required this.id,
    required this.sessionId,
    required this.messageType,
    required this.type,
    required this.level,
    required this.content,
    required this.createdAt,
    this.metadata = const <String, Object?>{},
    this.commandId,
    this.commandType,
    this.commandStatus,
  });

  final String id;
  final String sessionId;
  final String messageType;
  final LogEventType type;
  final LogLevel level;
  final String content;
  final DateTime createdAt;
  final Map<String, Object?> metadata;
  final String? commandId;
  final String? commandType;
  final String? commandStatus;
}
