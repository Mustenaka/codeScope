enum PromptCommandTaskStatus {
  pending,
  running,
  success,
  failed;

  bool get isTerminal =>
      this == PromptCommandTaskStatus.success ||
      this == PromptCommandTaskStatus.failed;
}

class PromptCommandTask {
  const PromptCommandTask({
    required this.id,
    required this.sessionId,
    required this.taskType,
    required this.prompt,
    required this.status,
    required this.result,
    required this.createdAt,
    required this.updatedAt,
  });

  final String id;
  final String sessionId;
  final String taskType;
  final String prompt;
  final PromptCommandTaskStatus status;
  final String result;
  final DateTime createdAt;
  final DateTime updatedAt;

  PromptCommandTask copyWith({
    PromptCommandTaskStatus? status,
    String? result,
    DateTime? updatedAt,
  }) {
    return PromptCommandTask(
      id: id,
      sessionId: sessionId,
      taskType: taskType,
      prompt: prompt,
      status: status ?? this.status,
      result: result ?? this.result,
      createdAt: createdAt,
      updatedAt: updatedAt ?? this.updatedAt,
    );
  }
}
