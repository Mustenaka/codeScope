import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:intl/intl.dart' as intl;

import 'app_localizations_en.dart';
import 'app_localizations_zh.dart';

// ignore_for_file: type=lint

/// Callers can lookup localized strings with an instance of AppLocalizations
/// returned by `AppLocalizations.of(context)`.
///
/// Applications need to include `AppLocalizations.delegate()` in their app's
/// `localizationDelegates` list, and the locales they support in the app's
/// `supportedLocales` list. For example:
///
/// ```dart
/// import 'generated/app_localizations.dart';
///
/// return MaterialApp(
///   localizationsDelegates: AppLocalizations.localizationsDelegates,
///   supportedLocales: AppLocalizations.supportedLocales,
///   home: MyApplicationHome(),
/// );
/// ```
///
/// ## Update pubspec.yaml
///
/// Please make sure to update your pubspec.yaml to include the following
/// packages:
///
/// ```yaml
/// dependencies:
///   # Internationalization support.
///   flutter_localizations:
///     sdk: flutter
///   intl: any # Use the pinned version from flutter_localizations
///
///   # Rest of dependencies
/// ```
///
/// ## iOS Applications
///
/// iOS applications define key application metadata, including supported
/// locales, in an Info.plist file that is built into the application bundle.
/// To configure the locales supported by your app, you’ll need to edit this
/// file.
///
/// First, open your project’s ios/Runner.xcworkspace Xcode workspace file.
/// Then, in the Project Navigator, open the Info.plist file under the Runner
/// project’s Runner folder.
///
/// Next, select the Information Property List item, select Add Item from the
/// Editor menu, then select Localizations from the pop-up menu.
///
/// Select and expand the newly-created Localizations item then, for each
/// locale your application supports, add a new item and select the locale
/// you wish to add from the pop-up menu in the Value field. This list should
/// be consistent with the languages listed in the AppLocalizations.supportedLocales
/// property.
abstract class AppLocalizations {
  AppLocalizations(String locale)
      : localeName = intl.Intl.canonicalizedLocale(locale.toString());

  final String localeName;

  static AppLocalizations? of(BuildContext context) {
    return Localizations.of<AppLocalizations>(context, AppLocalizations);
  }

  static const LocalizationsDelegate<AppLocalizations> delegate =
      _AppLocalizationsDelegate();

  /// A list of this localizations delegate along with the default localizations
  /// delegates.
  ///
  /// Returns a list of localizations delegates containing this delegate along with
  /// GlobalMaterialLocalizations.delegate, GlobalCupertinoLocalizations.delegate,
  /// and GlobalWidgetsLocalizations.delegate.
  ///
  /// Additional delegates can be added by appending to this list in
  /// MaterialApp. This list does not have to be used at all if a custom list
  /// of delegates is preferred or required.
  static const List<LocalizationsDelegate<dynamic>> localizationsDelegates =
      <LocalizationsDelegate<dynamic>>[
    delegate,
    GlobalMaterialLocalizations.delegate,
    GlobalCupertinoLocalizations.delegate,
    GlobalWidgetsLocalizations.delegate,
  ];

  /// A list of this localizations delegate's supported locales.
  static const List<Locale> supportedLocales = <Locale>[
    Locale('en'),
    Locale('zh')
  ];

  /// No description provided for @appTitle.
  ///
  /// In en, this message translates to:
  /// **'codeScope Mobile'**
  String get appTitle;

  /// No description provided for @sessionsTitle.
  ///
  /// In en, this message translates to:
  /// **'codeScope Sessions'**
  String get sessionsTitle;

  /// No description provided for @settingsTitle.
  ///
  /// In en, this message translates to:
  /// **'Settings'**
  String get settingsTitle;

  /// No description provided for @settingsTooltip.
  ///
  /// In en, this message translates to:
  /// **'Settings'**
  String get settingsTooltip;

  /// No description provided for @remoteSessionsHeadline.
  ///
  /// In en, this message translates to:
  /// **'Remote AI coding sessions'**
  String get remoteSessionsHeadline;

  /// No description provided for @sessionsReady.
  ///
  /// In en, this message translates to:
  /// **'{count} sessions ready'**
  String sessionsReady(int count);

  /// No description provided for @mockTransportSuffix.
  ///
  /// In en, this message translates to:
  /// **'mock transport enabled'**
  String get mockTransportSuffix;

  /// No description provided for @retry.
  ///
  /// In en, this message translates to:
  /// **'Retry'**
  String get retry;

  /// No description provided for @files.
  ///
  /// In en, this message translates to:
  /// **'Files'**
  String get files;

  /// No description provided for @prompt.
  ///
  /// In en, this message translates to:
  /// **'Prompt'**
  String get prompt;

  /// No description provided for @sessionDetailsTitle.
  ///
  /// In en, this message translates to:
  /// **'Session details'**
  String get sessionDetailsTitle;

  /// No description provided for @fileBrowserPlaceholderTitle.
  ///
  /// In en, this message translates to:
  /// **'File browser placeholder'**
  String get fileBrowserPlaceholderTitle;

  /// No description provided for @fileBrowserPlaceholderBody.
  ///
  /// In en, this message translates to:
  /// **'Reserved for GET /api/sessions/:id/files/tree and file content APIs.'**
  String get fileBrowserPlaceholderBody;

  /// No description provided for @promptPlaceholderTitle.
  ///
  /// In en, this message translates to:
  /// **'Prompt dispatch placeholder'**
  String get promptPlaceholderTitle;

  /// No description provided for @promptPlaceholderBody.
  ///
  /// In en, this message translates to:
  /// **'Reserved for GET /api/prompts and POST /api/sessions/:id/commands/prompt.'**
  String get promptPlaceholderBody;

  /// No description provided for @currentSessionLabel.
  ///
  /// In en, this message translates to:
  /// **'Current session: {sessionId}'**
  String currentSessionLabel(String sessionId);

  /// No description provided for @settingsSectionGeneral.
  ///
  /// In en, this message translates to:
  /// **'General'**
  String get settingsSectionGeneral;

  /// No description provided for @settingsSectionConnection.
  ///
  /// In en, this message translates to:
  /// **'Connection'**
  String get settingsSectionConnection;

  /// No description provided for @languageLabel.
  ///
  /// In en, this message translates to:
  /// **'Language'**
  String get languageLabel;

  /// No description provided for @languageChinese.
  ///
  /// In en, this message translates to:
  /// **'Chinese'**
  String get languageChinese;

  /// No description provided for @languageEnglish.
  ///
  /// In en, this message translates to:
  /// **'English'**
  String get languageEnglish;

  /// No description provided for @dataSourceLabel.
  ///
  /// In en, this message translates to:
  /// **'Data source'**
  String get dataSourceLabel;

  /// No description provided for @dataSourceMock.
  ///
  /// In en, this message translates to:
  /// **'Mock'**
  String get dataSourceMock;

  /// No description provided for @dataSourceServer.
  ///
  /// In en, this message translates to:
  /// **'Server'**
  String get dataSourceServer;

  /// No description provided for @apiBaseUrlLabel.
  ///
  /// In en, this message translates to:
  /// **'REST API Base URL'**
  String get apiBaseUrlLabel;

  /// No description provided for @webSocketUrlLabel.
  ///
  /// In en, this message translates to:
  /// **'WebSocket URL'**
  String get webSocketUrlLabel;

  /// No description provided for @settingsHintRuntimeOnly.
  ///
  /// In en, this message translates to:
  /// **'Changes apply to the current run only.'**
  String get settingsHintRuntimeOnly;

  /// No description provided for @settingsTestConnection.
  ///
  /// In en, this message translates to:
  /// **'Test connection'**
  String get settingsTestConnection;

  /// No description provided for @settingsTestingConnection.
  ///
  /// In en, this message translates to:
  /// **'Testing connection...'**
  String get settingsTestingConnection;

  /// No description provided for @settingsConnectionSuccess.
  ///
  /// In en, this message translates to:
  /// **'Connection successful'**
  String get settingsConnectionSuccess;

  /// No description provided for @settingsConnectionMockOnly.
  ///
  /// In en, this message translates to:
  /// **'Mock mode does not require a server connection test.'**
  String get settingsConnectionMockOnly;

  /// No description provided for @settingsConnectionFailure.
  ///
  /// In en, this message translates to:
  /// **'Connection failed: {message}'**
  String settingsConnectionFailure(String message);

  /// No description provided for @saveSettings.
  ///
  /// In en, this message translates to:
  /// **'Save settings'**
  String get saveSettings;

  /// No description provided for @settingsSaved.
  ///
  /// In en, this message translates to:
  /// **'Settings applied'**
  String get settingsSaved;

  /// No description provided for @machineLabel.
  ///
  /// In en, this message translates to:
  /// **'Machine'**
  String get machineLabel;

  /// No description provided for @workspaceLabel.
  ///
  /// In en, this message translates to:
  /// **'Workspace'**
  String get workspaceLabel;

  /// No description provided for @statusLabel.
  ///
  /// In en, this message translates to:
  /// **'Status'**
  String get statusLabel;

  /// No description provided for @createdStatus.
  ///
  /// In en, this message translates to:
  /// **'Created'**
  String get createdStatus;

  /// No description provided for @runningStatus.
  ///
  /// In en, this message translates to:
  /// **'Running'**
  String get runningStatus;

  /// No description provided for @stoppedStatus.
  ///
  /// In en, this message translates to:
  /// **'Stopped'**
  String get stoppedStatus;

  /// No description provided for @failedStatus.
  ///
  /// In en, this message translates to:
  /// **'Failed'**
  String get failedStatus;

  /// No description provided for @logTypeAi.
  ///
  /// In en, this message translates to:
  /// **'AI'**
  String get logTypeAi;

  /// No description provided for @logTypeTerminal.
  ///
  /// In en, this message translates to:
  /// **'Terminal'**
  String get logTypeTerminal;

  /// No description provided for @logTypeCommand.
  ///
  /// In en, this message translates to:
  /// **'Command'**
  String get logTypeCommand;

  /// No description provided for @logTypeFile.
  ///
  /// In en, this message translates to:
  /// **'File'**
  String get logTypeFile;

  /// No description provided for @logTypeDiff.
  ///
  /// In en, this message translates to:
  /// **'Diff'**
  String get logTypeDiff;

  /// No description provided for @logTypeError.
  ///
  /// In en, this message translates to:
  /// **'Error'**
  String get logTypeError;

  /// No description provided for @logTypeHeartbeat.
  ///
  /// In en, this message translates to:
  /// **'Heartbeat'**
  String get logTypeHeartbeat;

  /// No description provided for @loadingErrorPrefix.
  ///
  /// In en, this message translates to:
  /// **'Load failed: {message}'**
  String loadingErrorPrefix(String message);

  /// No description provided for @settingsSectionAbout.
  ///
  /// In en, this message translates to:
  /// **'About'**
  String get settingsSectionAbout;

  /// No description provided for @appVersionLabel.
  ///
  /// In en, this message translates to:
  /// **'Version'**
  String get appVersionLabel;

  /// No description provided for @appBuildLabel.
  ///
  /// In en, this message translates to:
  /// **'Build'**
  String get appBuildLabel;
}

class _AppLocalizationsDelegate
    extends LocalizationsDelegate<AppLocalizations> {
  const _AppLocalizationsDelegate();

  @override
  Future<AppLocalizations> load(Locale locale) {
    return SynchronousFuture<AppLocalizations>(lookupAppLocalizations(locale));
  }

  @override
  bool isSupported(Locale locale) =>
      <String>['en', 'zh'].contains(locale.languageCode);

  @override
  bool shouldReload(_AppLocalizationsDelegate old) => false;
}

AppLocalizations lookupAppLocalizations(Locale locale) {
  // Lookup logic when only language code is specified.
  switch (locale.languageCode) {
    case 'en':
      return AppLocalizationsEn();
    case 'zh':
      return AppLocalizationsZh();
  }

  throw FlutterError(
      'AppLocalizations.delegate failed to load unsupported locale "$locale". This is likely '
      'an issue with the localizations generation tool. Please file an issue '
      'on GitHub with a reproducible sample app and the gen-l10n configuration '
      'that was used.');
}
