import '../app/app_environment.dart';
import 'connection_tester.dart';
import 'default_connection_tester.dart';
import 'mock/mock_data_provider.dart';
import 'mock/mock_rest_client.dart';
import 'mock/mock_websocket_client.dart';
import 'real/server_rest_client.dart';
import 'real/server_websocket_client.dart';
import 'rest_client.dart';
import 'websocket_client.dart';

class AppServices {
  AppServices({
    required this.environment,
    required this.restClient,
    required this.webSocketClient,
    required this.connectionTester,
  });

  final AppEnvironment environment;
  final CodeScopeRestClient restClient;
  final CodeScopeWebSocketClient webSocketClient;
  final CodeScopeConnectionTester connectionTester;

  factory AppServices.bootstrap(AppEnvironment environment) {
    final connectionTester = DefaultCodeScopeConnectionTester();
    if (environment.useMock) {
      final provider = MockDataProvider();
      return AppServices(
        environment: environment,
        restClient: MockCodeScopeRestClient(provider),
        webSocketClient: MockCodeScopeWebSocketClient(provider),
        connectionTester: connectionTester,
      );
    }

    return AppServices(
      environment: environment,
      restClient: ServerCodeScopeRestClient(environment.apiBaseUrl),
      webSocketClient: ServerCodeScopeWebSocketClient(
        environment.webSocketUrl,
      ),
      connectionTester: connectionTester,
    );
  }
}
