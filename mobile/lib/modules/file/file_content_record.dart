class FileContentRecord {
  const FileContentRecord({
    required this.path,
    required this.size,
    required this.previewable,
    this.reason,
    this.content,
    this.language,
  });

  final String path;
  final int size;
  final bool previewable;
  final String? reason;
  final String? content;
  final String? language;
}
