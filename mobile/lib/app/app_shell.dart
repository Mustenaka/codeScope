import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import '../l10n/generated/app_localizations.dart';

import '../modules/message/thread_detail_page.dart';
import '../modules/project/project_page.dart';
import '../modules/project/project_record.dart';
import '../modules/thread/project_detail_page.dart';
import '../modules/thread/thread_record.dart';
import '../modules/log/log_page.dart';
import '../modules/settings/settings_page.dart';
import '../modules/session/session_page.dart';
import '../services/app_services.dart';
import 'app_environment.dart';
import 'app_routes.dart';
import 'app_settings.dart';
import 'app_scope.dart';
import 'app_theme.dart';

class CodeScopeApp extends StatefulWidget {
  const CodeScopeApp({
    required this.initialEnvironment,
    super.key,
  });

  final AppEnvironment initialEnvironment;

  @override
  State<CodeScopeApp> createState() => _CodeScopeAppState();
}

class _CodeScopeAppState extends State<CodeScopeApp> {
  late final AppSettingsController _settingsController;
  late AppServices _services;

  @override
  void initState() {
    super.initState();
    _settingsController = AppSettingsController(widget.initialEnvironment)
      ..addListener(_handleSettingsChanged);
    _services =
        AppServices.bootstrap(_settingsController.state.toEnvironment());
  }

  @override
  void dispose() {
    _settingsController
      ..removeListener(_handleSettingsChanged)
      ..dispose();
    super.dispose();
  }

  void _handleSettingsChanged() {
    setState(() {
      _services =
          AppServices.bootstrap(_settingsController.state.toEnvironment());
    });
  }

  @override
  Widget build(BuildContext context) {
    return AppScope(
      services: _services,
      settingsController: _settingsController,
      child: MaterialApp(
        onGenerateTitle: (BuildContext context) {
          return AppLocalizations.of(context)!.appTitle;
        },
        debugShowCheckedModeBanner: false,
        theme: buildCodeScopeTheme(),
        locale: _settingsController.state.locale,
        supportedLocales: AppLocalizations.supportedLocales,
        localizationsDelegates: const <LocalizationsDelegate<dynamic>>[
          AppLocalizations.delegate,
          GlobalMaterialLocalizations.delegate,
          GlobalWidgetsLocalizations.delegate,
          GlobalCupertinoLocalizations.delegate,
        ],
        initialRoute: AppRoutes.projects,
        onGenerateRoute: (RouteSettings settings) {
          switch (settings.name) {
            case AppRoutes.projects:
              return MaterialPageRoute<void>(
                builder: (_) => const ProjectPage(),
                settings: settings,
              );
            case AppRoutes.projectDetail:
              final project = settings.arguments as ProjectRecord;
              return MaterialPageRoute<void>(
                builder: (_) => ProjectDetailPage(project: project),
                settings: settings,
              );
            case AppRoutes.threadDetail:
              final thread = settings.arguments as ThreadRecord;
              return MaterialPageRoute<void>(
                builder: (_) => ThreadDetailPage(thread: thread),
                settings: settings,
              );
            case AppRoutes.sessions:
              return MaterialPageRoute<void>(
                builder: (_) => const SessionPage(),
                settings: settings,
              );
            case AppRoutes.sessionDetail:
              final sessionId = settings.arguments as String;
              return MaterialPageRoute<void>(
                builder: (_) => LogPage(sessionId: sessionId),
                settings: settings,
              );
            case AppRoutes.settings:
              return MaterialPageRoute<void>(
                builder: (_) => const SettingsPage(),
                settings: settings,
              );
            default:
              return MaterialPageRoute<void>(
                builder: (_) => const ProjectPage(),
                settings: settings,
              );
          }
        },
      ),
    );
  }
}
