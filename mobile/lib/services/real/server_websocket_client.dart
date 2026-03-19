import 'dart:async';
import 'dart:convert';
import 'dart:io';

import '../../modules/log/log_event.dart';
import '../../modules/message/thread_message_record.dart';
import '../websocket_client.dart';
import 'server_api_mapper.dart';
import 'server_rest_client.dart';

typedef WebSocketConnector = Future<WebSocket> Function(Uri uri);

class ServerCodeScopeWebSocketClient implements CodeScopeWebSocketClient {
  ServerCodeScopeWebSocketClient(
    this.url, {
    WebSocketConnector? connector,
  }) : _connector = connector ?? _defaultConnector;

  final String url;
  final WebSocketConnector _connector;

  static Future<WebSocket> _defaultConnector(Uri uri) {
    return WebSocket.connect(uri.toString());
  }

  @override
  Stream<LogEvent> subscribeToProjects() async* {
    final socket = await _connect();

    try {
      await for (final Object? message in socket) {
        yield _mapMessage(message);
      }
    } on SocketException catch (error) {
      throw ServerApiException('WebSocket connection failed: ${error.message}');
    } on WebSocketException catch (error) {
      throw ServerApiException('WebSocket connection failed: ${error.message}');
    } finally {
      await socket.close();
    }
  }

  @override
  Stream<LogEvent> subscribeToProject(String projectId) async* {
    final socket = await _connect(projectId: projectId);

    try {
      await for (final Object? message in socket) {
        yield _mapMessage(message);
      }
    } on SocketException catch (error) {
      throw ServerApiException('WebSocket connection failed: ${error.message}');
    } on WebSocketException catch (error) {
      throw ServerApiException('WebSocket connection failed: ${error.message}');
    } finally {
      await socket.close();
    }
  }

  @override
  Stream<LogEvent> subscribeToSession(String sessionId) async* {
    final socket = await _connect(sessionId: sessionId);

    try {
      await for (final Object? message in socket) {
        yield _mapMessage(message);
      }
    } on SocketException catch (error) {
      throw ServerApiException('WebSocket connection failed: ${error.message}');
    } on WebSocketException catch (error) {
      throw ServerApiException('WebSocket connection failed: ${error.message}');
    } finally {
      await socket.close();
    }
  }

  @override
  Stream<ThreadMessageRecord> subscribeToThread(String threadId) async* {
    final socket = await _connect(threadId: threadId);

    try {
      await for (final Object? message in socket) {
        final mapped = _mapThreadMessage(message, threadId);
        if (mapped != null) {
          yield mapped;
        }
      }
    } on SocketException catch (error) {
      throw ServerApiException('WebSocket connection failed: ${error.message}');
    } on WebSocketException catch (error) {
      throw ServerApiException('WebSocket connection failed: ${error.message}');
    } finally {
      await socket.close();
    }
  }

  Future<WebSocket> _connect({
    String? sessionId,
    String? threadId,
    String? projectId,
  }) async {
    final uri = _buildUri(
      sessionId: sessionId,
      threadId: threadId,
      projectId: projectId,
    );
    try {
      final socket = await _connector(uri);
      socket.pingInterval = const Duration(seconds: 20);
      return socket;
    } on SocketException catch (error) {
      throw ServerApiException('WebSocket connection failed: ${error.message}');
    } on WebSocketException catch (error) {
      throw ServerApiException('WebSocket connection failed: ${error.message}');
    }
  }

  LogEvent _mapMessage(Object? message) {
    final raw = switch (message) {
      String value => value,
      List<int> value => utf8.decode(value),
      _ => throw const FormatException('Unsupported WebSocket frame type.'),
    };

    final decoded = jsonDecode(raw);
    if (decoded is! Map) {
      throw const FormatException('Expected a JSON object from WebSocket.');
    }

    return ServerApiMapper.logEventFromJson(
      decoded.map(
        (Object? key, Object? value) => MapEntry(key.toString(), value),
      ),
    );
  }

  ThreadMessageRecord? _mapThreadMessage(Object? message, String threadId) {
    final raw = switch (message) {
      String value => value,
      List<int> value => utf8.decode(value),
      _ => throw const FormatException('Unsupported WebSocket frame type.'),
    };
    final decoded = jsonDecode(raw);
    if (decoded is! Map) {
      throw const FormatException('Expected a JSON object from WebSocket.');
    }
    return ServerApiMapper.threadMessageFromEventJson(
      decoded.map(
        (Object? key, Object? value) => MapEntry(key.toString(), value),
      ),
      fallbackThreadId: threadId,
    );
  }

  Uri _buildUri({String? sessionId, String? threadId, String? projectId}) {
    final normalizedBase = url.endsWith('/') ? url : '$url/';
    final baseUri = Uri.parse(normalizedBase);
    final query = <String, String>{...baseUri.queryParameters};
    if (sessionId != null) {
      query['session_id'] = sessionId;
    }
    if (threadId != null) {
      query['thread_id'] = threadId;
    }
    if (projectId != null) {
      query['project_id'] = projectId;
    }
    return baseUri.replace(
      path: baseUri.path.endsWith('/')
          ? baseUri.path.substring(0, baseUri.path.length - 1)
          : baseUri.path,
      queryParameters: query,
    );
  }
}
