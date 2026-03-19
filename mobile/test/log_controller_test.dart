import 'package:codescope_mobile/services/mock/mock_data_provider.dart';
import 'package:codescope_mobile/services/mock/mock_rest_client.dart';
import 'package:codescope_mobile/services/mock/mock_websocket_client.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:codescope_mobile/modules/log/log_controller.dart';

void main() {
  test('loads historical events before subscribing to live stream', () async {
    final provider = MockDataProvider();
    final controller = LogController(
      restClient: MockCodeScopeRestClient(provider),
      webSocketClient: MockCodeScopeWebSocketClient(provider),
    );

    await controller.load('session-001');

    expect(controller.session?.id, 'session-001');
    expect(controller.events.isNotEmpty, true);

    controller.dispose();
  });
}
