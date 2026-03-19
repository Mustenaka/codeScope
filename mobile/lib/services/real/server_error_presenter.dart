String presentServerError(Object error) {
  final raw = error.toString();
  final normalized = raw.toLowerCase();

  if (normalized.contains('project has no writable bridge session') ||
      normalized.contains('thread has no writable bridge session')) {
    return 'No writable local agent is online for this workspace. Open Codex or Claude in the same workspace and wait for the bridge to reconnect, then try again.';
  }
  if (normalized.contains('bridge not connected')) {
    return 'The local agent bridge is offline right now. Reopen the workspace in Codex or Claude, wait for it to reconnect, then try again.';
  }
  return raw;
}
