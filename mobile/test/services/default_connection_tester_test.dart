import 'dart:convert';
import 'dart:io';

import 'package:codescope_mobile/app/app_environment.dart';
import 'package:codescope_mobile/services/connection_tester.dart';
import 'package:codescope_mobile/services/default_connection_tester.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  group('DefaultCodeScopeConnectionTester', () {
    late HttpServer server;
    late DefaultCodeScopeConnectionTester tester;

    setUp(() async {
      server = await HttpServer.bind(InternetAddress.loopbackIPv4, 0);
      tester = DefaultCodeScopeConnectionTester();
    });

    tearDown(() async {
      await server.close(force: true);
    });

    test('verifies REST health and websocket handshake in server mode',
        () async {
      server.listen((HttpRequest request) async {
        if (request.uri.path == '/api/health') {
          request.response.headers.contentType = ContentType.json;
          request.response.write(jsonEncode(<String, Object?>{'status': 'ok'}));
          await request.response.close();
          return;
        }

        if (request.uri.path == '/ws/mobile') {
          expect(
              request.uri.queryParameters['session_id'], 'connectivity-check');
          final socket = await WebSocketTransformer.upgrade(request);
          await socket.close();
          return;
        }

        request.response.statusCode = HttpStatus.notFound;
        await request.response.close();
      });

      final result = await tester.test(
        AppEnvironment(
          useMock: false,
          apiBaseUrl: 'http://${server.address.host}:${server.port}/api',
          webSocketUrl: 'ws://${server.address.host}:${server.port}/ws/mobile',
        ),
      );

      expect(result.status, ConnectionTestStatus.success);
    });
  });
}
