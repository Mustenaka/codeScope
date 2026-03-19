import 'package:codescope_mobile/app/app_environment.dart';
import 'package:codescope_mobile/services/app_services.dart';
import 'package:codescope_mobile/services/mock/mock_rest_client.dart';
import 'package:codescope_mobile/services/mock/mock_websocket_client.dart';
import 'package:codescope_mobile/services/real/server_rest_client.dart';
import 'package:codescope_mobile/services/real/server_websocket_client.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('bootstrap returns mock clients when environment uses mock', () {
    final services = AppServices.bootstrap(AppEnvironment.mock);

    expect(services.restClient, isA<MockCodeScopeRestClient>());
    expect(services.webSocketClient, isA<MockCodeScopeWebSocketClient>());
  });

  test('bootstrap returns real clients when environment uses server', () {
    const environment = AppEnvironment(
      useMock: false,
      apiBaseUrl: 'http://127.0.0.1:8080/api',
      webSocketUrl: 'ws://127.0.0.1:8080/ws/mobile',
    );

    final services = AppServices.bootstrap(environment);

    expect(services.restClient, isA<ServerCodeScopeRestClient>());
    expect(services.webSocketClient, isA<ServerCodeScopeWebSocketClient>());
  });
}
