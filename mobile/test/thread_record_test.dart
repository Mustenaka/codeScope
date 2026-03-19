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
}
