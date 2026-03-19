class ProjectRecord {
  const ProjectRecord({
    required this.id,
    required this.name,
    required this.workspaceRoot,
    required this.machineId,
    required this.threadCount,
    required this.runningThreadCount,
    required this.createdAt,
    required this.lastActivityAt,
  });

  final String id;
  final String name;
  final String workspaceRoot;
  final String machineId;
  final int threadCount;
  final int runningThreadCount;
  final DateTime createdAt;
  final DateTime lastActivityAt;
}
