// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for English (`en`).
class AppLocalizationsEn extends AppLocalizations {
  AppLocalizationsEn([String locale = 'en']) : super(locale);

  @override
  String get appTitle => 'codeScope Mobile';

  @override
  String get sessionsTitle => 'codeScope Sessions';

  @override
  String get settingsTitle => 'Settings';

  @override
  String get settingsTooltip => 'Settings';

  @override
  String get remoteSessionsHeadline => 'Remote AI coding sessions';

  @override
  String sessionsReady(int count) {
    return '$count sessions ready';
  }

  @override
  String get mockTransportSuffix => 'mock transport enabled';

  @override
  String get retry => 'Retry';

  @override
  String get files => 'Files';

  @override
  String get prompt => 'Prompt';

  @override
  String get sessionDetailsTitle => 'Session details';

  @override
  String get fileBrowserPlaceholderTitle => 'File browser placeholder';

  @override
  String get fileBrowserPlaceholderBody =>
      'Reserved for GET /api/sessions/:id/files/tree and file content APIs.';

  @override
  String get promptPlaceholderTitle => 'Prompt dispatch placeholder';

  @override
  String get promptPlaceholderBody =>
      'Reserved for GET /api/prompts and POST /api/sessions/:id/commands/prompt.';

  @override
  String currentSessionLabel(String sessionId) {
    return 'Current session: $sessionId';
  }

  @override
  String get settingsSectionGeneral => 'General';

  @override
  String get settingsSectionConnection => 'Connection';

  @override
  String get languageLabel => 'Language';

  @override
  String get languageChinese => 'Chinese';

  @override
  String get languageEnglish => 'English';

  @override
  String get dataSourceLabel => 'Data source';

  @override
  String get dataSourceMock => 'Mock';

  @override
  String get dataSourceServer => 'Server';

  @override
  String get apiBaseUrlLabel => 'REST API Base URL';

  @override
  String get webSocketUrlLabel => 'WebSocket URL';

  @override
  String get settingsHintRuntimeOnly =>
      'Changes apply to the current run only.';

  @override
  String get settingsTestConnection => 'Test connection';

  @override
  String get settingsTestingConnection => 'Testing connection...';

  @override
  String get settingsConnectionSuccess => 'Connection successful';

  @override
  String get settingsConnectionMockOnly =>
      'Mock mode does not require a server connection test.';

  @override
  String settingsConnectionFailure(String message) {
    return 'Connection failed: $message';
  }

  @override
  String get saveSettings => 'Save settings';

  @override
  String get settingsSaved => 'Settings applied';

  @override
  String get machineLabel => 'Machine';

  @override
  String get workspaceLabel => 'Workspace';

  @override
  String get statusLabel => 'Status';

  @override
  String get createdStatus => 'Created';

  @override
  String get runningStatus => 'Running';

  @override
  String get stoppedStatus => 'Stopped';

  @override
  String get failedStatus => 'Failed';

  @override
  String get logTypeAi => 'AI';

  @override
  String get logTypeTerminal => 'Terminal';

  @override
  String get logTypeCommand => 'Command';

  @override
  String get logTypeFile => 'File';

  @override
  String get logTypeDiff => 'Diff';

  @override
  String get logTypeError => 'Error';

  @override
  String get logTypeHeartbeat => 'Heartbeat';

  @override
  String loadingErrorPrefix(String message) {
    return 'Load failed: $message';
  }

  @override
  String get settingsSectionAbout => 'About';

  @override
  String get appVersionLabel => 'Version';

  @override
  String get appBuildLabel => 'Build';
}
