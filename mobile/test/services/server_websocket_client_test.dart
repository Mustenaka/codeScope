import 'dart:convert';
import 'dart:io';

import 'package:codescope_mobile/modules/log/log_event.dart';
import 'package:codescope_mobile/modules/message/thread_message_record.dart';
import 'package:codescope_mobile/services/real/server_websocket_client.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  group('ServerCodeScopeWebSocketClient', () {
    late HttpServer server;
    late ServerCodeScopeWebSocketClient client;

    setUp(() async {
      server = await HttpServer.bind(InternetAddress.loopbackIPv4, 0);
      client = ServerCodeScopeWebSocketClient(
        'ws://${server.address.host}:${server.port}/ws/mobile',
      );
    });

    tearDown(() async {
      await server.close(force: true);
    });

    test('subscribes with session_id and maps server records into log events',
        () async {
      server.listen((HttpRequest request) async {
        expect(request.uri.path, '/ws/mobile');
        expect(request.uri.queryParameters['session_id'], 'session-001');

        final socket = await WebSocketTransformer.upgrade(request);
        socket.add(
          jsonEncode(<String, Object?>{
            'id': 'event-001',
            'message_id': 'msg-001',
            'session_id': 'session-001',
            'message_type': 'event',
            'event_type': 'ai_output',
            'timestamp': '2026-03-17T09:00:00Z',
            'created_at': '2026-03-17T09:00:00Z',
            'payload': <String, Object?>{
              'content': 'Patch applied.',
              'level': 'warning',
            },
          }),
        );
        await socket.close();
      });

      final event = await client.subscribeToSession('session-001').first;

      expect(event.sessionId, 'session-001');
      expect(event.type, LogEventType.aiOutput);
      expect(event.level, LogLevel.warning);
      expect(event.content, 'Patch applied.');
    });

    test(
        'subscribes with thread_id and maps event records into thread messages',
        () async {
      server.listen((HttpRequest request) async {
        expect(request.uri.path, '/ws/mobile');
        expect(request.uri.queryParameters['thread_id'], 'thread-001');

        final socket = await WebSocketTransformer.upgrade(request);
        socket.add(
          jsonEncode(<String, Object?>{
            'id': 'event-002',
            'message_id': 'msg-002',
            'session_id': 'session-001',
            'message_type': 'event',
            'event_type': 'ai_output',
            'timestamp': '2026-03-19T09:05:00Z',
            'created_at': '2026-03-19T09:05:00Z',
            'payload': <String, Object?>{
              'thread_id': 'thread-001',
              'content': 'Applied follow-up patch',
            },
          }),
        );
        await socket.close();
      });

      final message = await client.subscribeToThread('thread-001').first;

      expect(message.threadId, 'thread-001');
      expect(message.role, ThreadMessageRole.assistant);
      expect(message.content, 'Applied follow-up patch');
    });

    test('subscribes with thread_id and maps real captured user events',
        () async {
      server.listen((HttpRequest request) async {
        expect(request.uri.path, '/ws/mobile');
        expect(request.uri.queryParameters['thread_id'], 'thread-001');

        final socket = await WebSocketTransformer.upgrade(request);
        socket.add(
          jsonEncode(<String, Object?>{
            'id': 'event-004',
            'message_id': 'msg-004',
            'session_id': 'session-001',
            'message_type': 'event',
            'event_type': 'command',
            'timestamp': '2026-03-19T09:06:00Z',
            'created_at': '2026-03-19T09:06:00Z',
            'payload': <String, Object?>{
              'thread_id': 'thread-001',
              'role': 'user',
              'content': '请同步新的需求指令。',
            },
          }),
        );
        await socket.close();
      });

      final message = await client.subscribeToThread('thread-001').first;

      expect(message.threadId, 'thread-001');
      expect(message.role, ThreadMessageRole.user);
      expect(message.content, '请同步新的需求指令。');
    });

    test('subscribes with project_id and maps server records into log events',
        () async {
      server.listen((HttpRequest request) async {
        expect(request.uri.path, '/ws/mobile');
        expect(request.uri.queryParameters['project_id'], 'project-001');

        final socket = await WebSocketTransformer.upgrade(request);
        socket.add(
          jsonEncode(<String, Object?>{
            'id': 'event-003',
            'message_id': 'msg-003',
            'session_id': 'session-001',
            'message_type': 'event',
            'event_type': 'terminal_output',
            'timestamp': '2026-03-19T09:10:00Z',
            'created_at': '2026-03-19T09:10:00Z',
            'payload': <String, Object?>{
              'project_id': 'project-001',
              'content': 'Refreshing project list',
            },
          }),
        );
        await socket.close();
      });

      final event = await client.subscribeToProject('project-001').first;

      expect(event.sessionId, 'session-001');
      expect(event.type, LogEventType.terminalOutput);
      expect(event.content, 'Refreshing project list');
      expect(event.metadata['project_id'], 'project-001');
    });
  });
}
