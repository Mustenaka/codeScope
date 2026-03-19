import 'package:codescope_mobile/app/app_environment.dart';
import 'package:codescope_mobile/app/app_scope.dart';
import 'package:codescope_mobile/app/app_settings.dart';
import 'package:codescope_mobile/app/app_shell.dart';
import 'package:codescope_mobile/l10n/generated/app_localizations.dart';
import 'package:codescope_mobile/modules/settings/settings_page.dart';
import 'package:codescope_mobile/services/app_services.dart';
import 'package:codescope_mobile/services/connection_tester.dart';
import 'package:codescope_mobile/services/mock/mock_data_provider.dart';
import 'package:codescope_mobile/services/mock/mock_rest_client.dart';
import 'package:codescope_mobile/services/mock/mock_websocket_client.dart';
import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  testWidgets('opens settings and keeps chinese as the default locale', (
    WidgetTester tester,
  ) async {
    await tester.pumpWidget(
      const CodeScopeApp(
        initialEnvironment: AppEnvironment.mock,
      ),
    );

    await tester.pumpAndSettle();
    expect(find.byIcon(Icons.settings_rounded), findsOneWidget);
    expect(find.text('codeScope'), findsOneWidget);

    await tester.tap(find.byIcon(Icons.settings_rounded));
    await tester.pumpAndSettle();

    expect(find.byType(SegmentedButton<AppLanguage>), findsOneWidget);
    expect(find.text('中文'), findsOneWidget);
  });

  testWidgets('tests server connectivity from settings before saving', (
    WidgetTester tester,
  ) async {
    final settingsController = AppSettingsController(AppEnvironment.mock);
    final services = AppServices(
      environment: AppEnvironment.mock,
      restClient: MockCodeScopeRestClient(MockDataProvider()),
      webSocketClient: MockCodeScopeWebSocketClient(MockDataProvider()),
      connectionTester: const _FakeConnectionTester(),
    );

    await tester.pumpWidget(
      AppScope(
        services: services,
        settingsController: settingsController,
        child: const MaterialApp(
          localizationsDelegates: <LocalizationsDelegate<dynamic>>[
            AppLocalizations.delegate,
            GlobalMaterialLocalizations.delegate,
            GlobalWidgetsLocalizations.delegate,
            GlobalCupertinoLocalizations.delegate,
          ],
          supportedLocales: AppLocalizations.supportedLocales,
          home: SettingsPage(),
        ),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('Server'));
    await tester.pumpAndSettle();

    final fields = find.byType(TextField);
    expect(fields, findsNWidgets(2));
    await tester.enterText(fields.at(0), 'http://127.0.0.1:8080/api');
    await tester.enterText(fields.at(1), 'ws://127.0.0.1:8080/ws/mobile');

    final testButton = find.byKey(const ValueKey('settings-test-connection'));
    await tester.scrollUntilVisible(
      testButton,
      200,
      scrollable: find.byType(Scrollable).first,
    );
    await tester.ensureVisible(testButton);
    await tester.pumpAndSettle();
    await tester.tap(testButton);
    await tester.pump();
    await tester.pumpAndSettle();

    expect(find.byType(SnackBar), findsOneWidget);
  });
}

class _FakeConnectionTester implements CodeScopeConnectionTester {
  const _FakeConnectionTester();

  @override
  Future<ConnectionTestResult> test(AppEnvironment environment) async {
    return const ConnectionTestResult(
      status: ConnectionTestStatus.success,
      message: 'Connection successful',
    );
  }
}
