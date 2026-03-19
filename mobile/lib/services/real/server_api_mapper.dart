import 'dart:convert';

import '../../modules/file/file_content_record.dart';
import '../../modules/file/file_tree_node.dart';
import '../../modules/log/log_event.dart';
import '../../modules/message/thread_message_record.dart';
import '../../modules/project/project_record.dart';
import '../../modules/prompt/prompt_command_task.dart';
import '../../modules/session/session_record.dart';
import '../../modules/thread/thread_record.dart';

class ServerApiMapper {
  const ServerApiMapper._();

  static SessionRecord sessionFromJson(Map<String, Object?> json) {
    return SessionRecord(
      id: _readString(json, 'id'),
      projectName: _readString(json, 'project_name'),
      workspaceRoot: _readString(json, 'workspace_root'),
      machineId: _readString(json, 'machine_id'),
      status: _sessionStatusFromJson(_readString(json, 'status')),
      startedAt: _readDateTime(json, 'started_at'),
      updatedAt: _readDateTime(json, 'updated_at'),
      endedAt: _readOptionalDateTime(json, 'ended_at'),
      summary: '',
    );
  }

  static ProjectRecord projectFromJson(Map<String, Object?> json) {
    return ProjectRecord(
      id: _readString(json, 'id'),
      name: _readString(json, 'name'),
      workspaceRoot: _readString(json, 'workspace_root'),
      machineId: _readString(json, 'machine_id'),
      threadCount: _readInt(json, 'thread_count'),
      runningThreadCount: _readInt(json, 'running_thread_count'),
      lastActivityAt: _readDateTime(json, 'last_activity_at'),
    );
  }

  static ThreadRecord threadFromJson(Map<String, Object?> json) {
    return ThreadRecord(
      id: _readString(json, 'id'),
      projectId: _readString(json, 'project_id'),
      sessionId: _readString(json, 'session_id'),
      title: _readString(json, 'title'),
      agentKind: _readNullableString(json, 'agent_kind'),
      status: _threadStatusFromJson(_readString(json, 'status')),
      summary: _readNullableString(json, 'summary') ?? '',
      lastActivityAt: _readDateTime(json, 'last_activity_at'),
      startedAt: _readDateTime(json, 'started_at'),
      endedAt: _readOptionalDateTime(json, 'ended_at'),
    );
  }

  static ThreadMessageRecord threadMessageFromJson(Map<String, Object?> json) {
    return ThreadMessageRecord(
      id: _readString(json, 'id'),
      threadId: _readString(json, 'thread_id'),
      role: _threadMessageRoleFromJson(_readString(json, 'role')),
      content: _readString(json, 'content'),
      createdAt: _readDateTime(json, 'created_at'),
      sequence: _readInt(json, 'sequence'),
      sourceType: _readNullableString(json, 'source_type'),
    );
  }

  static ThreadMessageRecord? threadMessageFromEventJson(
    Map<String, Object?> json, {
    required String fallbackThreadId,
  }) {
    final payload = _readPayload(json);
    final messageType = _readString(json, 'message_type');
    final eventType = _readNullableString(json, 'event_type');
    final threadId =
        _readNullableString(payload, 'thread_id') ?? fallbackThreadId;

    if (messageType == 'command' &&
        _readNullableString(json, 'command_type') == 'send_prompt') {
      final content = _readNullableString(payload, 'content');
      if (content == null || content.isEmpty) {
        return null;
      }
      return ThreadMessageRecord(
        id: _readString(json, 'id'),
        threadId: threadId,
        role: ThreadMessageRole.user,
        content: content,
        createdAt: _readDateTime(json, 'created_at', fallbackKey: 'timestamp'),
        sequence: 0,
        sourceType: 'command',
      );
    }

    if (messageType == 'event' &&
        eventType == 'command' &&
        _readNullableString(payload, 'role') == 'user') {
      final content = _readNullableString(payload, 'content');
      if (content == null || content.isEmpty) {
        return null;
      }
      return ThreadMessageRecord(
        id: _readString(json, 'id'),
        threadId: threadId,
        role: ThreadMessageRole.user,
        content: content,
        createdAt: _readDateTime(json, 'created_at', fallbackKey: 'timestamp'),
        sequence: 0,
        sourceType: eventType,
      );
    }

    if (messageType == 'event' && eventType == 'ai_output') {
      final content = _readNullableString(payload, 'content');
      if (content == null || content.isEmpty) {
        return null;
      }
      return ThreadMessageRecord(
        id: _readString(json, 'id'),
        threadId: threadId,
        role: ThreadMessageRole.assistant,
        content: content,
        createdAt: _readDateTime(json, 'created_at', fallbackKey: 'timestamp'),
        sequence: 0,
        sourceType: eventType,
      );
    }

    return null;
  }

  static LogEvent logEventFromJson(Map<String, Object?> json) {
    final payload = _readPayload(json);
    return LogEvent(
      id: _readString(json, 'id'),
      sessionId: _readString(json, 'session_id'),
      messageType: _readString(json, 'message_type'),
      type: _logEventTypeFromJson(
        messageType: _readString(json, 'message_type'),
        eventType: _readNullableString(json, 'event_type'),
      ),
      level: _logLevelFromJson(
        level: _readNullableString(payload, 'level'),
        eventType: _readNullableString(json, 'event_type'),
      ),
      content: _contentFromPayload(payload),
      createdAt: _readDateTime(json, 'created_at', fallbackKey: 'timestamp'),
      metadata: payload,
      commandId: _readNullableString(json, 'command_id'),
      commandType: _readNullableString(json, 'command_type'),
      commandStatus: _readNullableString(json, 'status'),
    );
  }

  static List<SessionRecord> sessionListFromJson(List<Object?> json) {
    return json
        .map((Object? item) => sessionFromJson(_asMap(item)))
        .toList(growable: false);
  }

  static List<ProjectRecord> projectListFromJson(List<Object?> json) {
    return json
        .map((Object? item) => projectFromJson(_asMap(item)))
        .toList(growable: false);
  }

  static List<ThreadRecord> threadListFromJson(List<Object?> json) {
    return json
        .map((Object? item) => threadFromJson(_asMap(item)))
        .toList(growable: false);
  }

  static List<ThreadMessageRecord> threadMessageListFromJson(
    List<Object?> json,
  ) {
    return json
        .map((Object? item) => threadMessageFromJson(_asMap(item)))
        .toList(growable: false);
  }

  static List<LogEvent> logEventListFromJson(List<Object?> json) {
    return json
        .map((Object? item) => logEventFromJson(_asMap(item)))
        .toList(growable: false);
  }

  static List<PromptCommandTask> commandTaskListFromJson(List<Object?> json) {
    return json
        .map((Object? item) => commandTaskFromJson(_asMap(item)))
        .toList(growable: false);
  }

  static PromptCommandTask commandTaskFromJson(Map<String, Object?> json) {
    final payload = _readPayload(json);
    return PromptCommandTask(
      id: _readString(json, 'id'),
      sessionId: _readString(json, 'session_id'),
      taskType: _readString(json, 'task_type'),
      prompt: _readNullableString(payload, 'content') ?? '',
      status: _commandTaskStatusFromJson(_readString(json, 'status')),
      result: _readNullableString(json, 'result') ?? '',
      createdAt: _readDateTime(json, 'created_at'),
      updatedAt: _readDateTime(json, 'updated_at'),
    );
  }

  static List<FileTreeNode> fileTreeFromJson(List<Object?> json) {
    return json
        .map((Object? item) => fileTreeNodeFromJson(_asMap(item)))
        .toList(growable: false);
  }

  static FileTreeNode fileTreeNodeFromJson(Map<String, Object?> json) {
    final childrenRaw = json['children'];
    final children = childrenRaw is List<Object?>
        ? childrenRaw
            .map((Object? item) => fileTreeNodeFromJson(_asMap(item)))
            .toList(growable: false)
        : const <FileTreeNode>[];

    return FileTreeNode(
      name: _readString(json, 'name'),
      path: _readString(json, 'path'),
      type: _readString(json, 'type'),
      size: _readNullableInt(json, 'size'),
      previewable: _readBool(json, 'previewable', fallback: false),
      children: children,
    );
  }

  static FileContentRecord fileContentFromJson(Map<String, Object?> json) {
    return FileContentRecord(
      path: _readString(json, 'path'),
      size: _readInt(json, 'size'),
      previewable: _readBool(json, 'previewable', fallback: false),
      reason: _readNullableString(json, 'reason'),
      content: _readNullableString(json, 'content'),
      language: _readNullableString(json, 'language'),
    );
  }

  static Map<String, Object?> errorFromJson(Object? json) => _asMap(json);

  static Map<String, Object?> _asMap(Object? value) {
    if (value is Map<String, Object?>) {
      return value;
    }
    if (value is Map) {
      return value.map(
        (Object? key, Object? entryValue) =>
            MapEntry(key.toString(), entryValue),
      );
    }
    throw const FormatException('Expected a JSON object.');
  }

  static Map<String, Object?> _readPayload(Map<String, Object?> json) {
    final payload = json['payload'];
    if (payload == null) {
      return const <String, Object?>{};
    }
    return _asMap(payload);
  }

  static String _contentFromPayload(Map<String, Object?> payload) {
    final content = payload['content'];
    if (content is String) {
      return content;
    }
    if (payload['error'] case final String error when error.isNotEmpty) {
      return error;
    }
    if (payload['result'] case final String result when result.isNotEmpty) {
      return result;
    }
    if (payload.isEmpty) {
      return '';
    }
    return jsonEncode(payload);
  }

  static SessionStatus _sessionStatusFromJson(String value) {
    switch (value) {
      case 'created':
        return SessionStatus.created;
      case 'running':
        return SessionStatus.running;
      case 'stopped':
        return SessionStatus.stopped;
      case 'failed':
        return SessionStatus.failed;
    }
    throw FormatException('Unsupported session status "$value".');
  }

  static PromptCommandTaskStatus _commandTaskStatusFromJson(String value) {
    switch (value) {
      case 'pending':
        return PromptCommandTaskStatus.pending;
      case 'running':
        return PromptCommandTaskStatus.running;
      case 'success':
        return PromptCommandTaskStatus.success;
      case 'failed':
        return PromptCommandTaskStatus.failed;
    }
    throw FormatException('Unsupported command task status "$value".');
  }

  static ThreadStatus _threadStatusFromJson(String value) {
    switch (value) {
      case 'running':
        return ThreadStatus.running;
      case 'waiting_prompt':
        return ThreadStatus.waitingPrompt;
      case 'waiting_review':
        return ThreadStatus.waitingReview;
      case 'completed':
        return ThreadStatus.completed;
      case 'blocked':
        return ThreadStatus.blocked;
    }
    throw FormatException('Unsupported thread status "$value".');
  }

  static ThreadMessageRole _threadMessageRoleFromJson(String value) {
    switch (value) {
      case 'user':
        return ThreadMessageRole.user;
      case 'assistant':
        return ThreadMessageRole.assistant;
      case 'system':
        return ThreadMessageRole.system;
    }
    throw FormatException('Unsupported thread message role "$value".');
  }

  static LogEventType _logEventTypeFromJson({
    required String messageType,
    required String? eventType,
  }) {
    if (messageType == 'heartbeat') {
      return LogEventType.heartbeat;
    }
    if (messageType == 'command_result') {
      return LogEventType.commandResult;
    }

    switch (eventType) {
      case 'ai_output':
        return LogEventType.aiOutput;
      case 'terminal_output':
        return LogEventType.terminalOutput;
      case 'command':
        return LogEventType.command;
      case 'file_change':
        return LogEventType.fileChange;
      case 'diff':
        return LogEventType.diff;
      case 'error':
        return LogEventType.error;
      case 'heartbeat':
        return LogEventType.heartbeat;
    }

    throw FormatException(
      'Unsupported event type "$eventType" for message_type "$messageType".',
    );
  }

  static LogLevel _logLevelFromJson({
    required String? level,
    required String? eventType,
  }) {
    switch (level) {
      case 'debug':
        return LogLevel.debug;
      case 'info':
        return LogLevel.info;
      case 'warning':
        return LogLevel.warning;
      case 'error':
        return LogLevel.error;
      case null:
        if (eventType == 'error') {
          return LogLevel.error;
        }
        return LogLevel.info;
    }

    throw FormatException('Unsupported log level "$level".');
  }

  static DateTime _readDateTime(
    Map<String, Object?> json,
    String key, {
    String? fallbackKey,
  }) {
    final primary = _readNullableString(json, key);
    if (primary != null && primary.isNotEmpty) {
      return DateTime.parse(primary);
    }

    if (fallbackKey != null) {
      final fallback = _readNullableString(json, fallbackKey);
      if (fallback != null && fallback.isNotEmpty) {
        return DateTime.parse(fallback);
      }
    }

    throw FormatException('Missing date field "$key".');
  }

  static DateTime? _readOptionalDateTime(
      Map<String, Object?> json, String key) {
    final value = _readNullableString(json, key);
    if (value == null || value.isEmpty) {
      return null;
    }
    return DateTime.parse(value);
  }

  static String _readString(Map<String, Object?> json, String key) {
    final value = json[key];
    if (value is String) {
      return value;
    }
    throw FormatException('Missing or invalid string field "$key".');
  }

  static String? _readNullableString(Map<String, Object?> json, String key) {
    final value = json[key];
    if (value == null) {
      return null;
    }
    if (value is String) {
      return value;
    }
    throw FormatException('Invalid string field "$key".');
  }

  static int _readInt(Map<String, Object?> json, String key) {
    final value = json[key];
    if (value is int) {
      return value;
    }
    throw FormatException('Missing or invalid int field "$key".');
  }

  static int? _readNullableInt(Map<String, Object?> json, String key) {
    final value = json[key];
    if (value == null) {
      return null;
    }
    if (value is int) {
      return value;
    }
    throw FormatException('Invalid int field "$key".');
  }

  static bool _readBool(
    Map<String, Object?> json,
    String key, {
    required bool fallback,
  }) {
    final value = json[key];
    if (value == null) {
      return fallback;
    }
    if (value is bool) {
      return value;
    }
    throw FormatException('Invalid bool field "$key".');
  }
}
