import 'package:codescope_mobile/services/real/server_error_presenter.dart';
import 'package:codescope_mobile/services/real/server_rest_client.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('maps no writable bridge session into actionable prompt guidance', () {
    final message = presentServerError(
      ServerApiException(
        'project has no writable bridge session',
        statusCode: 409,
      ),
    );

    expect(
      message,
      'No writable local agent is online for this workspace. Open Codex or Claude in the same workspace and wait for the bridge to reconnect, then try again.',
    );
  });
}
