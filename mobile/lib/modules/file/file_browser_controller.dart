import 'package:flutter/foundation.dart';

import '../../services/rest_client.dart';
import 'file_content_record.dart';
import 'file_tree_node.dart';

class FileBrowserController extends ChangeNotifier {
  FileBrowserController({
    required this.restClient,
    this.sessionId,
    this.projectId,
  }) : assert(
          (sessionId != null && sessionId != '') ||
              (projectId != null && projectId != ''),
          'FileBrowserController requires either sessionId or projectId.',
        );

  final CodeScopeRestClient restClient;
  final String? sessionId;
  final String? projectId;

  bool _loading = false;
  String? _errorMessage;
  List<FileTreeNode> _tree = const <FileTreeNode>[];
  FileContentRecord? _selectedContent;
  String? _selectedPath;

  bool get isLoading => _loading;
  String? get errorMessage => _errorMessage;
  List<FileTreeNode> get tree => _tree;
  FileContentRecord? get selectedContent => _selectedContent;
  String? get selectedPath => _selectedPath;

  Future<void> load() async {
    _loading = true;
    _errorMessage = null;
    notifyListeners();

    try {
      _tree = projectId != null
          ? await restClient.fetchProjectFileTree(projectId!)
          : await restClient.fetchSessionFileTree(sessionId!);
    } catch (error) {
      _errorMessage = error.toString();
    } finally {
      _loading = false;
      notifyListeners();
    }
  }

  Future<void> selectFile(String path) async {
    _selectedPath = path;
    _errorMessage = null;
    notifyListeners();

    try {
      _selectedContent = projectId != null
          ? await restClient.fetchProjectFileContent(projectId!, path)
          : await restClient.fetchSessionFileContent(sessionId!, path);
    } catch (error) {
      _errorMessage = error.toString();
      _selectedContent = null;
    } finally {
      notifyListeners();
    }
  }
}
