class FileTreeNode {
  const FileTreeNode({
    required this.name,
    required this.path,
    required this.type,
    this.size,
    this.previewable = false,
    this.children = const <FileTreeNode>[],
  });

  final String name;
  final String path;
  final String type;
  final int? size;
  final bool previewable;
  final List<FileTreeNode> children;

  bool get isDirectory => type == 'directory';
}
