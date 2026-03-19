import 'log_event.dart';

class LogEventViewModel {
  const LogEventViewModel(this.event);

  final LogEvent event;

  String get summary {
    switch (event.type) {
      case LogEventType.command:
        return _firstNonEmpty(<String?>[
          _stringValue(event.metadata, 'content'),
          _stringValue(event.metadata, 'command_line'),
          event.content,
        ]);
      case LogEventType.commandResult:
        return _firstNonEmpty(<String?>[
          _stringValue(event.metadata, 'result'),
          _stringValue(event.metadata, 'error'),
          event.content,
        ]);
      case LogEventType.fileChange:
        return _firstNonEmpty(<String?>[
          _joinNonEmpty(<String?>[
            _stringValue(event.metadata, 'op'),
            _stringValue(event.metadata, 'path'),
          ], ' '),
          event.content,
        ]);
      default:
        return event.content;
    }
  }

  List<MapEntry<String, Object?>> get highlights {
    switch (event.type) {
      case LogEventType.command:
        return _compactEntries(<MapEntry<String, Object?>>[
          if (event.commandType != null)
            MapEntry<String, Object?>('type', event.commandType),
          if (event.commandId != null)
            MapEntry<String, Object?>('command_id', event.commandId),
          if (_stringValue(event.metadata, 'process_name') != null)
            MapEntry<String, Object?>(
                'process', _stringValue(event.metadata, 'process_name')),
          if (_stringValue(event.metadata, 'pid') != null)
            MapEntry<String, Object?>(
                'pid', _stringValue(event.metadata, 'pid')),
          if (_stringValue(event.metadata, 'source') != null)
            MapEntry<String, Object?>(
                'source', _stringValue(event.metadata, 'source')),
        ]);
      case LogEventType.commandResult:
        return _compactEntries(<MapEntry<String, Object?>>[
          if (event.commandStatus != null)
            MapEntry<String, Object?>('status', event.commandStatus),
          if (event.commandId != null)
            MapEntry<String, Object?>('command_id', event.commandId),
          if (_stringValue(event.metadata, 'accepted') != null)
            MapEntry<String, Object?>(
                'accepted', _stringValue(event.metadata, 'accepted')),
        ]);
      case LogEventType.fileChange:
        return _compactEntries(<MapEntry<String, Object?>>[
          if (_stringValue(event.metadata, 'path') != null)
            MapEntry<String, Object?>(
                'path', _stringValue(event.metadata, 'path')),
          if (_stringValue(event.metadata, 'op') != null)
            MapEntry<String, Object?>('op', _stringValue(event.metadata, 'op')),
        ]);
      default:
        return _compactEntries(event.metadata.entries.take(4).toList());
    }
  }

  List<MapEntry<String, Object?>> get rawMetadata =>
      event.metadata.entries.toList();

  static List<MapEntry<String, Object?>> _compactEntries(
      List<MapEntry<String, Object?>> entries) {
    return entries.where((MapEntry<String, Object?> entry) {
      final value = entry.value;
      return value != null && value.toString().isNotEmpty;
    }).toList();
  }

  static String _firstNonEmpty(List<String?> values) {
    for (final String? value in values) {
      if (value != null && value.isNotEmpty) {
        return value;
      }
    }
    return '';
  }

  static String? _joinNonEmpty(List<String?> values, String separator) {
    final joined = values
        .whereType<String>()
        .where((String value) => value.isNotEmpty)
        .join(separator);
    return joined.isEmpty ? null : joined;
  }

  static String? _stringValue(Map<String, Object?> metadata, String key) {
    final value = metadata[key];
    if (value == null) {
      return null;
    }
    return value.toString();
  }
}
