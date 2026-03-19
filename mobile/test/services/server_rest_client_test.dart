import 'dart:convert';
import 'dart:io';

import 'package:codescope_mobile/modules/log/log_event.dart';
import 'package:codescope_mobile/modules/message/thread_message_record.dart';
import 'package:codescope_mobile/modules/prompt/prompt_command_task.dart';
import 'package:codescope_mobile/modules/session/session_record.dart';
import 'package:codescope_mobile/modules/thread/thread_record.dart';
import 'package:codescope_mobile/services/real/server_rest_client.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  group('ServerCodeScopeRestClient', () {
    late HttpServer server;
    late ServerCodeScopeRestClient client;

    setUp(() async {
      server = await HttpServer.bind(InternetAddress.loopbackIPv4, 0);
      client = ServerCodeScopeRestClient(
        'http://${server.address.host}:${server.port}/api',
      );
    });

    tearDown(() async {
      await server.close(force: true);
    });

    test(
        'fetches sessions, session detail, events, files, and commands from server JSON',
        () async {
      server.listen((HttpRequest request) async {
        if (request.uri.path == '/api/sessions') {
          _writeJson(request.response, <Map<String, Object?>>[
            <String, Object?>{
              'id': 'session-001',
              'project_name': 'codeScope/mobile',
              'workspace_root': 'D:/Work/Code/Cross/codeScope/mobile',
              'machine_id': 'devbox',
              'status': 'running',
              'bridge_online': true,
              'started_at': '2026-03-17T08:00:00Z',
              'updated_at': '2026-03-17T08:05:00Z',
            },
          ]);
          return;
        }

        if (request.uri.path == '/api/sessions/session-001') {
          _writeJson(request.response, <String, Object?>{
            'id': 'session-001',
            'project_name': 'codeScope/mobile',
            'workspace_root': 'D:/Work/Code/Cross/codeScope/mobile',
            'machine_id': 'devbox',
            'status': 'running',
            'bridge_online': true,
            'started_at': '2026-03-17T08:00:00Z',
            'updated_at': '2026-03-17T08:05:00Z',
            'ended_at': null,
          });
          return;
        }

        if (request.uri.path == '/api/sessions/session-001/events') {
          _writeJson(request.response, <Map<String, Object?>>[
            <String, Object?>{
              'id': 'event-001',
              'message_id': 'msg-001',
              'session_id': 'session-001',
              'message_type': 'event',
              'event_type': 'terminal_output',
              'timestamp': '2026-03-17T08:01:00Z',
              'created_at': '2026-03-17T08:01:00Z',
              'payload': <String, Object?>{
                'content': 'flutter test',
                'level': 'info',
              },
            },
          ]);
          return;
        }

        if (request.uri.path == '/api/sessions/session-001/files/tree') {
          _writeJson(request.response, <Map<String, Object?>>[
            <String, Object?>{
              'name': 'lib',
              'path': 'lib',
              'type': 'directory',
              'children': <Map<String, Object?>>[
                <String, Object?>{
                  'name': 'main.dart',
                  'path': 'lib/main.dart',
                  'type': 'file',
                  'size': 128,
                  'previewable': true,
                },
              ],
            },
          ]);
          return;
        }

        if (request.uri.path == '/api/sessions/session-001/files/content') {
          _writeJson(request.response, <String, Object?>{
            'path': request.uri.queryParameters['path'],
            'size': 128,
            'previewable': true,
            'language': 'dart',
            'content': 'void main() {}',
          });
          return;
        }

        if (request.uri.path == '/api/sessions/session-001/commands' &&
            request.method == 'GET') {
          _writeJson(request.response, <Map<String, Object?>>[
            <String, Object?>{
              'id': 'cmd-001',
              'session_id': 'session-001',
              'task_type': 'send_prompt',
              'payload': <String, Object?>{
                'content': 'continue implementing',
              },
              'status': 'running',
              'result': '',
              'created_at': '2026-03-18T09:00:00Z',
              'updated_at': '2026-03-18T09:00:30Z',
            },
          ]);
          return;
        }

        if (request.uri.path == '/api/sessions/session-001/commands/prompt' &&
            request.method == 'POST') {
          final body = await utf8.decoder.bind(request).join();
          expect(body, contains('continue implementing'));
          _writeJson(
              request.response,
              <String, Object?>{
                'id': 'cmd-002',
                'session_id': 'session-001',
                'task_type': 'send_prompt',
                'payload': <String, Object?>{
                  'content': 'continue implementing',
                },
                'status': 'running',
                'result': '',
                'created_at': '2026-03-18T09:01:00Z',
                'updated_at': '2026-03-18T09:01:00Z',
              },
              statusCode: HttpStatus.created);
          return;
        }

        if (request.uri.path == '/api/projects') {
          _writeJson(request.response, <Map<String, Object?>>[
            <String, Object?>{
              'id': 'project-001',
              'name': 'codeScope',
              'workspace_root': 'D:/Work/Code/Cross/codeScope',
              'machine_id': 'devbox',
              'thread_count': 2,
              'running_thread_count': 1,
              'last_activity_at': '2026-03-19T10:05:00Z',
            },
          ]);
          return;
        }

        if (request.uri.path == '/api/projects/project-001') {
          _writeJson(request.response, <String, Object?>{
            'id': 'project-001',
            'name': 'codeScope',
            'workspace_root': 'D:/Work/Code/Cross/codeScope',
            'machine_id': 'devbox',
            'thread_count': 2,
            'running_thread_count': 1,
            'last_activity_at': '2026-03-19T10:05:00Z',
          });
          return;
        }

        if (request.uri.path == '/api/projects/project-001/threads') {
          _writeJson(request.response, <Map<String, Object?>>[
            <String, Object?>{
              'id': 'session-001',
              'project_id': 'project-001',
              'session_id': 'session-001',
              'title': 'Implemented server API',
              'agent_kind': 'codex',
              'status': 'running',
              'summary': 'Implemented server API',
              'last_activity_at': '2026-03-19T10:05:00Z',
              'started_at': '2026-03-19T10:00:00Z',
            },
          ]);
          return;
        }

        if (request.uri.path == '/api/threads/session-001') {
          _writeJson(request.response, <String, Object?>{
            'id': 'session-001',
            'project_id': 'project-001',
            'session_id': 'session-001',
            'title': 'Implemented server API',
            'agent_kind': 'codex',
            'status': 'running',
            'summary': 'Implemented server API',
            'last_activity_at': '2026-03-19T10:05:00Z',
            'started_at': '2026-03-19T10:00:00Z',
          });
          return;
        }

        if (request.uri.path == '/api/threads/session-001/messages') {
          _writeJson(request.response, <Map<String, Object?>>[
            <String, Object?>{
              'id': 'event-001',
              'thread_id': 'session-001',
              'role': 'user',
              'content': 'continue implementing',
              'created_at': '2026-03-19T10:01:00Z',
              'sequence': 1,
              'source_type': 'command',
            },
            <String, Object?>{
              'id': 'event-002',
              'thread_id': 'session-001',
              'role': 'assistant',
              'content': 'Implemented server API',
              'created_at': '2026-03-19T10:05:00Z',
              'sequence': 2,
              'source_type': 'ai_output',
            },
          ]);
          return;
        }

        request.response.statusCode = HttpStatus.notFound;
        await request.response.close();
      });

      final sessions = await client.fetchSessions();
      final detail = await client.fetchSessionDetail('session-001');
      final events = await client.fetchSessionEvents('session-001');
      final tree = await client.fetchSessionFileTree('session-001');
      final content = await client.fetchSessionFileContent(
        'session-001',
        'lib/main.dart',
      );
      final commands = await client.fetchSessionCommands('session-001');
      final created = await client.sendPrompt(
        'session-001',
        'continue implementing',
      );
      final projects = await client.fetchProjects();
      final project = await client.fetchProjectDetail('project-001');
      final threads = await client.fetchProjectThreads('project-001');
      final thread = await client.fetchThreadDetail('session-001');
      final messages = await client.fetchThreadMessages('session-001');

      expect(sessions, hasLength(1));
      expect(sessions.single.projectName, 'codeScope/mobile');
      expect(sessions.single.status, SessionStatus.running);
      expect(detail.id, 'session-001');
      expect(detail.startedAt, DateTime.parse('2026-03-17T08:00:00Z'));
      expect(events, hasLength(1));
      expect(events.single.type, LogEventType.terminalOutput);
      expect(events.single.level, LogLevel.info);
      expect(events.single.content, 'flutter test');
      expect(tree.single.children.single.path, 'lib/main.dart');
      expect(content.path, 'lib/main.dart');
      expect(content.previewable, isTrue);
      expect(commands.single.status, PromptCommandTaskStatus.running);
      expect(created.id, 'cmd-002');
      expect(projects.single.name, 'codeScope');
      expect(project.threadCount, 2);
      expect(threads.single.status, ThreadStatus.running);
      expect(thread.id, 'session-001');
      expect(messages, hasLength(2));
      expect(messages.last.role, ThreadMessageRole.assistant);
    });

    test('throws a readable error when server returns a failure response',
        () async {
      server.listen((HttpRequest request) async {
        _writeJson(
          request.response,
          <String, Object?>{'error': 'session store unavailable'},
          statusCode: HttpStatus.internalServerError,
        );
      });

      expect(
        client.fetchSessions(),
        throwsA(
          isA<ServerApiException>().having(
            (ServerApiException error) => error.message,
            'message',
            contains('session store unavailable'),
          ),
        ),
      );
    });
  });
}

void _writeJson(HttpResponse response, Object body,
    {int statusCode = HttpStatus.ok}) {
  response.statusCode = statusCode;
  response.headers.contentType = ContentType.json;
  response.write(jsonEncode(body));
  response.close();
}
