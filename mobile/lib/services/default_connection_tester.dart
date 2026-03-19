import 'dart:async';
import 'dart:convert';
import 'dart:io';

import '../app/app_environment.dart';
import 'connection_tester.dart';

typedef ConnectionHttpClientFactory = HttpClient Function();
typedef ConnectionWebSocketConnector = Future<WebSocket> Function(Uri uri);

class DefaultCodeScopeConnectionTester implements CodeScopeConnectionTester {
  DefaultCodeScopeConnectionTester({
    HttpClient Function()? httpClientFactory,
    ConnectionWebSocketConnector? webSocketConnector,
    Duration timeout = const Duration(seconds: 3),
  })  : _httpClientFactory = httpClientFactory ?? HttpClient.new,
        _webSocketConnector = webSocketConnector ??
            ((Uri uri) => WebSocket.connect(uri.toString())),
        _timeout = timeout;

  final ConnectionHttpClientFactory _httpClientFactory;
  final ConnectionWebSocketConnector _webSocketConnector;
  final Duration _timeout;

  @override
  Future<ConnectionTestResult> test(AppEnvironment environment) async {
    if (environment.useMock) {
      return const ConnectionTestResult(
        status: ConnectionTestStatus.skipped,
        message: 'Mock mode enabled.',
      );
    }

    try {
      await _probeHealth(environment.apiBaseUrl);
      await _probeWebSocket(environment.webSocketUrl);
      return const ConnectionTestResult(
        status: ConnectionTestStatus.success,
        message: 'Connection successful',
      );
    } on TimeoutException {
      return const ConnectionTestResult(
        status: ConnectionTestStatus.failure,
        message: 'Connection timed out',
      );
    } on SocketException catch (error) {
      return ConnectionTestResult(
        status: ConnectionTestStatus.failure,
        message: error.message,
      );
    } on WebSocketException catch (error) {
      return ConnectionTestResult(
        status: ConnectionTestStatus.failure,
        message: error.message,
      );
    } on HttpException catch (error) {
      return ConnectionTestResult(
        status: ConnectionTestStatus.failure,
        message: error.message,
      );
    } on FormatException catch (error) {
      return ConnectionTestResult(
        status: ConnectionTestStatus.failure,
        message: error.message,
      );
    }
  }

  Future<void> _probeHealth(String apiBaseUrl) async {
    final client = _httpClientFactory();
    client.connectionTimeout = _timeout;

    try {
      final request =
          await client.getUrl(_buildHealthUri(apiBaseUrl)).timeout(_timeout);
      request.headers.set(HttpHeaders.acceptHeader, ContentType.json.mimeType);

      final response = await request.close().timeout(_timeout);
      final body = await utf8.decoder.bind(response).join().timeout(_timeout);

      if (response.statusCode < 200 || response.statusCode >= 300) {
        throw HttpException('Health check failed with ${response.statusCode}');
      }

      if (body.isNotEmpty) {
        final decoded = jsonDecode(body);
        if (decoded is! Map || decoded['status'] != 'ok') {
          throw const FormatException('Unexpected health response.');
        }
      }
    } finally {
      client.close(force: true);
    }
  }

  Future<void> _probeWebSocket(String webSocketUrl) async {
    final socket = await _webSocketConnector(
      _buildWebSocketUri(webSocketUrl),
    ).timeout(_timeout);
    await socket.close();
  }

  Uri _buildHealthUri(String apiBaseUrl) {
    final normalizedBase =
        apiBaseUrl.endsWith('/') ? apiBaseUrl : '$apiBaseUrl/';
    return Uri.parse(normalizedBase).resolve('health');
  }

  Uri _buildWebSocketUri(String webSocketUrl) {
    final baseUri = Uri.parse(webSocketUrl);
    return baseUri.replace(
      queryParameters: <String, String>{
        ...baseUri.queryParameters,
        'session_id': 'connectivity-check',
      },
    );
  }
}
