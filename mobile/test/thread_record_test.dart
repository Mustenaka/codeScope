import 'package:codescope_mobile/modules/thread/thread_record.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('displaySummary returns captured summary when available', () {
    final record = ThreadRecord(
      id: 'thread-1',
      projectId: 'project-1',
      sessionId: 'session-1',
      title: 'Thread 1',
      status: ThreadStatus.running,
      summary: 'Implemented server API',
      lastActivityAt: DateTime.utc(2026, 3, 19, 10),
      startedAt: DateTime.utc(2026, 3, 19, 9),
    );

    expect(record.displaySummary, 'Implemented server API');
  });

  test('displaySummary explains empty running threads explicitly', () {
    final record = ThreadRecord(
      id: 'thread-1',
      projectId: 'project-1',
      sessionId: 'session-1',
      title: 'Thread 1',
      status: ThreadStatus.running,
      lastActivityAt: DateTime.utc(2026, 3, 19, 10),
      startedAt: DateTime.utc(2026, 3, 19, 9),
    );

    expect(
      record.displaySummary,
      'Running. No readable message summary has been captured yet.',
    );
  });

  test('displaySummary explains offline and stale threads explicitly', () {
    final offline = ThreadRecord(
      id: 'thread-offline-1',
      projectId: 'project-1',
      sessionId: 'session-1',
      title: 'Offline thread',
      status: ThreadStatus.offline,
      lastActivityAt: DateTime.utc(2026, 3, 19, 10),
      startedAt: DateTime.utc(2026, 3, 19, 9),
    );
    final stale = ThreadRecord(
      id: 'thread-stale-1',
      projectId: 'project-1',
      sessionId: 'session-1',
      title: 'Stale thread',
      status: ThreadStatus.stale,
      lastActivityAt: DateTime.utc(2026, 3, 19, 10),
      startedAt: DateTime.utc(2026, 3, 19, 9),
    );

    expect(
      offline.displaySummary,
      'Offline. The last known thread state is preserved until the bridge reconnects.',
    );
    expect(
      stale.displaySummary,
      'Stale. The thread has gone quiet longer than the active window.',
    );
  });
}
