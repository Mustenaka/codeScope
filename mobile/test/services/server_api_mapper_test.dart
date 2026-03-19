import 'package:codescope_mobile/modules/log/log_event.dart';
import 'package:codescope_mobile/modules/message/thread_message_record.dart';
import 'package:codescope_mobile/modules/prompt/prompt_command_task.dart';
import 'package:codescope_mobile/modules/thread/thread_record.dart';
import 'package:codescope_mobile/services/real/server_api_mapper.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  group('ServerApiMapper', () {
    test('maps command result records into command result log events', () {
      final event = ServerApiMapper.logEventFromJson(<String, Object?>{
        'id': 'event-002',
        'message_id': 'msg-002',
        'session_id': 'session-001',
        'message_type': 'command_result',
        'command_id': 'cmd-001',
        'command_type': 'send_prompt',
        'status': 'failed',
        'timestamp': '2026-03-17T08:02:00Z',
        'created_at': '2026-03-17T08:02:00Z',
        'payload': <String, Object?>{
          'accepted': false,
          'error': 'side-channel mode does not support prompt injection',
          'result': '',
        },
      });

      expect(event.type, LogEventType.commandResult);
      expect(event.commandId, 'cmd-001');
      expect(event.commandType, 'send_prompt');
      expect(event.commandStatus, 'failed');
      expect(
          event.content, 'side-channel mode does not support prompt injection');
    });

    test('maps recursive file tree nodes and file content payloads', () {
      final nodes = ServerApiMapper.fileTreeFromJson(<Object?>[
        <String, Object?>{
          'name': 'lib',
          'path': 'lib',
          'type': 'directory',
          'children': <Object?>[
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

      final content = ServerApiMapper.fileContentFromJson(<String, Object?>{
        'path': 'lib/main.dart',
        'size': 128,
        'previewable': true,
        'language': 'dart',
        'content': 'void main() {}',
      });

      expect(nodes, hasLength(1));
      expect(nodes.single.children, hasLength(1));
      expect(nodes.single.children.single.path, 'lib/main.dart');
      expect(content.path, 'lib/main.dart');
      expect(content.previewable, isTrue);
      expect(content.content, 'void main() {}');
    });

    test('maps prompt command task records', () {
      final task = ServerApiMapper.commandTaskFromJson(<String, Object?>{
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
      });

      expect(task.id, 'cmd-001');
      expect(task.prompt, 'continue implementing');
      expect(task.status, PromptCommandTaskStatus.running);
    });

    test('maps project, thread, and message payloads', () {
      final project = ServerApiMapper.projectFromJson(<String, Object?>{
        'id': 'project-1',
        'name': 'codeScope',
        'workspace_root': 'D:/Work/Code/Cross/codeScope',
        'machine_id': 'devbox',
        'thread_count': 2,
        'running_thread_count': 1,
        'last_activity_at': '2026-03-19T10:05:00Z',
      });
      final thread = ServerApiMapper.threadFromJson(<String, Object?>{
        'id': 'session-1',
        'project_id': 'project-1',
        'session_id': 'session-1',
        'title': 'Implemented server API',
        'agent_kind': 'codex',
        'status': 'running',
        'summary': 'Implemented server API',
        'last_activity_at': '2026-03-19T10:05:00Z',
        'started_at': '2026-03-19T10:00:00Z',
      });
      final message = ServerApiMapper.threadMessageFromJson(
        <String, Object?>{
          'id': 'event-1',
          'thread_id': 'session-1',
          'role': 'assistant',
          'content': 'Implemented server API',
          'created_at': '2026-03-19T10:05:00Z',
          'sequence': 1,
          'source_type': 'ai_output',
        },
      );

      expect(project.name, 'codeScope');
      expect(project.threadCount, 2);
      expect(thread.status, ThreadStatus.running);
      expect(thread.title, 'Implemented server API');
      expect(message.role, ThreadMessageRole.assistant);
      expect(message.content, 'Implemented server API');
    });

    test('maps real captured user thread events into thread messages', () {
      final message = ServerApiMapper.threadMessageFromEventJson(
        <String, Object?>{
          'id': 'event-user-1',
          'message_type': 'event',
          'event_type': 'command',
          'timestamp': '2026-03-19T10:06:00Z',
          'created_at': '2026-03-19T10:06:00Z',
          'payload': <String, Object?>{
            'thread_id': 'thread-001',
            'role': 'user',
            'content': '请继续修复 bridge 采集链路。',
          },
        },
        fallbackThreadId: 'thread-001',
      );

      expect(message, isNotNull);
      expect(message!.role, ThreadMessageRole.user);
      expect(message.content, '请继续修复 bridge 采集链路。');
      expect(message.sourceType, 'command');
    });
  });
}
