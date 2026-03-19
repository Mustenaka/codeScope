import '../l10n/generated/app_localizations.dart';

import '../modules/log/log_event.dart';
import '../modules/session/session_record.dart';

extension SessionStatusLocalizationX on SessionStatus {
  String localized(AppLocalizations l10n) {
    switch (this) {
      case SessionStatus.created:
        return l10n.createdStatus;
      case SessionStatus.running:
        return l10n.runningStatus;
      case SessionStatus.stopped:
        return l10n.stoppedStatus;
      case SessionStatus.failed:
        return l10n.failedStatus;
    }
  }
}

extension LogEventTypeLocalizationX on LogEventType {
  String localized(AppLocalizations l10n) {
    switch (this) {
      case LogEventType.aiOutput:
        return l10n.logTypeAi;
      case LogEventType.terminalOutput:
        return l10n.logTypeTerminal;
      case LogEventType.command:
        return l10n.logTypeCommand;
      case LogEventType.commandResult:
        return l10n.logTypeCommand;
      case LogEventType.fileChange:
        return l10n.logTypeFile;
      case LogEventType.diff:
        return l10n.logTypeDiff;
      case LogEventType.error:
        return l10n.logTypeError;
      case LogEventType.heartbeat:
        return l10n.logTypeHeartbeat;
    }
  }
}
