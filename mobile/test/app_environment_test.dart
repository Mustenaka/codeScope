import 'package:codescope_mobile/app/app_environment.dart';
import 'package:codescope_mobile/app/app_settings.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  group('AppEnvironment.fromValues', () {
    test('defaults to mock when no overrides are provided', () {
      final environment = AppEnvironment.fromValues();

      expect(environment.useMock, isTrue);
      expect(environment.apiBaseUrl, 'http://192.168.31.137:8080/api');
      expect(environment.webSocketUrl, 'ws://192.168.31.137:8080/ws/mobile');
    });

    test('switches to server mode when URLs are provided without useMock', () {
      final environment = AppEnvironment.fromValues(
        apiBaseUrlValue: 'http://127.0.0.1:8080/api',
        webSocketUrlValue: 'ws://127.0.0.1:8080/ws/mobile',
      );

      expect(environment.useMock, isFalse);
      expect(environment.apiBaseUrl, 'http://127.0.0.1:8080/api');
      expect(environment.webSocketUrl, 'ws://127.0.0.1:8080/ws/mobile');
    });

    test('respects an explicit mock override from dart-define values', () {
      final environment = AppEnvironment.fromValues(
        useMockValue: 'true',
        apiBaseUrlValue: 'http://127.0.0.1:8080/api',
        webSocketUrlValue: 'ws://127.0.0.1:8080/ws/mobile',
      );

      expect(environment.useMock, isTrue);
    });

    test('defaults app settings locale to chinese', () {
      final state =
          AppSettingsState.fromEnvironment(AppEnvironment.fromValues());

      expect(state.language, AppLanguage.chinese);
    });
  });
}
