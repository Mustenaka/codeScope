import 'package:codescope_mobile/modules/log/log_event.dart';
import 'package:codescope_mobile/modules/session/session_record.dart';
import 'package:codescope_mobile/modules/session/session_signal_analyzer.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('derives waiting prompt state and summary from latest meaningful event',
      () {
    final session = SessionRecord(
      id: 'session-001',
      projectName: 'codeScope',
      workspaceRoot: 'D:/Work/Code/Cross/codeScope',
      machineId: 'devbox',
      status: SessionStatus.running,
      startedAt: DateTime.parse('2026-03-18T09:00:00Z'),
      updatedAt: DateTime.parse('2026-03-18T09:05:00Z'),
    );

    final digest = SessionSignalAnalyzer.analyze(
      session,
      <LogEvent>[
        LogEvent(
          id: 'event-heartbeat',
          sessionId: 'session-001',
          messageType: 'heartbeat',
          type: LogEventType.heartbeat,
          level: LogLevel.debug,
          content: 'Bridge heartbeat acknowledged.',
          createdAt: DateTime.parse('2026-03-18T09:04:00Z'),
        ),
        LogEvent(
          id: 'event-ai',
          sessionId: 'session-001',
          messageType: 'event',
          type: LogEventType.aiOutput,
          level: LogLevel.info,
          content:
              'Implementation finished, waiting for next prompt from user.',
          createdAt: DateTime.parse('2026-03-18T09:05:00Z'),
        ),
      ],
    );

    expect(digest.state, AgentUserState.waitingPrompt);
    expect(digest.summary,
        'Implementation finished, waiting for next prompt from user.');
    expect(digest.visibleEvents, hasLength(1));
  });

  test('derives waiting review state from review/check keywords', () {
    final session = SessionRecord(
      id: 'session-002',
      projectName: 'codeScope',
      workspaceRoot: 'D:/Work/Code/Cross/codeScope',
      machineId: 'devbox',
      status: SessionStatus.running,
      startedAt: DateTime.parse('2026-03-18T08:00:00Z'),
      updatedAt: DateTime.parse('2026-03-18T08:30:00Z'),
    );

    final digest = SessionSignalAnalyzer.analyze(
      session,
      <LogEvent>[
        LogEvent(
          id: 'event-review',
          sessionId: 'session-002',
          messageType: 'event',
          type: LogEventType.aiOutput,
          level: LogLevel.info,
          content:
              'All changes are ready, please confirm and check before merge.',
          createdAt: DateTime.parse('2026-03-18T08:30:00Z'),
        ),
      ],
    );

    expect(digest.state, AgentUserState.waitingReview);
  });
}
