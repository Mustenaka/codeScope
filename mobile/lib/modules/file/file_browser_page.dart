import 'package:flutter/material.dart';

import '../../app/app_scope.dart';
import '../../l10n/generated/app_localizations.dart';
import '../settings/settings_button.dart';
import 'file_browser_controller.dart';
import 'file_content_record.dart';
import 'file_tree_node.dart';

class FileBrowserPage extends StatefulWidget {
  const FileBrowserPage({
    this.sessionId,
    this.projectId,
    super.key,
  }) : assert(
          (sessionId != null && sessionId != '') ||
              (projectId != null && projectId != ''),
          'FileBrowserPage requires either sessionId or projectId.',
        );

  final String? sessionId;
  final String? projectId;

  @override
  State<FileBrowserPage> createState() => _FileBrowserPageState();
}

class _FileBrowserPageState extends State<FileBrowserPage> {
  FileBrowserController? _controller;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    _controller ??= FileBrowserController(
      restClient: AppScope.servicesOf(context).restClient,
      sessionId: widget.sessionId,
      projectId: widget.projectId,
    )..load();
  }

  @override
  Widget build(BuildContext context) {
    final controller = _controller;
    if (controller == null) {
      return const SizedBox.shrink();
    }

    final l10n = AppLocalizations.of(context)!;
    return AnimatedBuilder(
      animation: controller,
      builder: (BuildContext context, Widget? child) {
        return Scaffold(
          appBar: AppBar(
            title: Text(l10n.files),
            actions: const <Widget>[SettingsButton()],
          ),
          body: controller.isLoading && controller.tree.isEmpty
              ? const Center(child: CircularProgressIndicator())
              : controller.errorMessage != null && controller.tree.isEmpty
                  ? Center(child: Text(controller.errorMessage!))
                  : LayoutBuilder(
                      builder:
                          (BuildContext context, BoxConstraints constraints) {
                        final isNarrow = constraints.maxWidth < 900;
                        if (isNarrow) {
                          return Column(
                            children: <Widget>[
                              Expanded(
                                child: _FileTreePane(
                                  tree: controller.tree,
                                  selectedPath: controller.selectedPath,
                                  onFileSelected: controller.selectFile,
                                ),
                              ),
                              SizedBox(
                                height: constraints.maxHeight * 0.45,
                                child: _FilePreviewPane(
                                  sessionId: widget.sessionId,
                                  projectId: widget.projectId,
                                  content: controller.selectedContent,
                                ),
                              ),
                            ],
                          );
                        }

                        return Row(
                          children: <Widget>[
                            Expanded(
                              child: _FileTreePane(
                                tree: controller.tree,
                                selectedPath: controller.selectedPath,
                                onFileSelected: controller.selectFile,
                              ),
                            ),
                            Expanded(
                              child: _FilePreviewPane(
                                sessionId: widget.sessionId,
                                projectId: widget.projectId,
                                content: controller.selectedContent,
                              ),
                            ),
                          ],
                        );
                      },
                    ),
        );
      },
    );
  }
}

class _FileTreePane extends StatelessWidget {
  const _FileTreePane({
    required this.tree,
    required this.selectedPath,
    required this.onFileSelected,
  });

  final List<FileTreeNode> tree;
  final String? selectedPath;
  final Future<void> Function(String path) onFileSelected;

  @override
  Widget build(BuildContext context) {
    return ListView(
      padding: const EdgeInsets.all(12),
      children: <Widget>[
        for (final FileTreeNode node in tree)
          _FileNodeView(
            node: node,
            selectedPath: selectedPath,
            onFileSelected: onFileSelected,
          ),
      ],
    );
  }
}

class _FileNodeView extends StatelessWidget {
  const _FileNodeView({
    required this.node,
    required this.selectedPath,
    required this.onFileSelected,
  });

  final FileTreeNode node;
  final String? selectedPath;
  final Future<void> Function(String path) onFileSelected;

  @override
  Widget build(BuildContext context) {
    if (node.isDirectory) {
      return ExpansionTile(
        leading: const Icon(Icons.folder_rounded),
        title: Text(node.name),
        children: node.children
            .map(
              (FileTreeNode child) => _FileNodeView(
                node: child,
                selectedPath: selectedPath,
                onFileSelected: onFileSelected,
              ),
            )
            .toList(),
      );
    }

    return ListTile(
      dense: true,
      selected: node.path == selectedPath,
      leading: Icon(
        node.previewable
            ? Icons.description_rounded
            : Icons.insert_drive_file_outlined,
      ),
      title: Text(node.name),
      subtitle: Text(node.path),
      onTap: () => onFileSelected(node.path),
    );
  }
}

class _FilePreviewPane extends StatelessWidget {
  const _FilePreviewPane({
    required this.sessionId,
    required this.projectId,
    required this.content,
  });

  final String? sessionId;
  final String? projectId;
  final FileContentRecord? content;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Container(
      decoration: BoxDecoration(
        border: Border(
          left: BorderSide(color: theme.colorScheme.outlineVariant),
          top: BorderSide(color: theme.colorScheme.outlineVariant),
        ),
      ),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: content == null
            ? Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: <Widget>[
                  Text(
                    'Select a file to preview',
                    style: theme.textTheme.titleMedium?.copyWith(
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                  const SizedBox(height: 8),
                  Text(projectId != null
                      ? 'Project: $projectId'
                      : 'Session: $sessionId'),
                ],
              )
            : Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: <Widget>[
                  Text(
                    content!.path,
                    style: theme.textTheme.titleMedium?.copyWith(
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                  const SizedBox(height: 8),
                  Text('Size: ${content!.size} bytes'),
                  if ((content!.language ?? '').isNotEmpty)
                    Text('Language: ${content!.language}'),
                  const SizedBox(height: 12),
                  Expanded(
                    child: DecoratedBox(
                      decoration: BoxDecoration(
                        color: theme.colorScheme.surfaceContainerHighest,
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: Padding(
                        padding: const EdgeInsets.all(12),
                        child: SingleChildScrollView(
                          child: Text(
                            content!.previewable
                                ? (content!.content ?? '')
                                : 'Preview unavailable: ${content!.reason ?? 'unknown'}',
                            style: theme.textTheme.bodySmall?.copyWith(
                              height: 1.5,
                            ),
                          ),
                        ),
                      ),
                    ),
                  ),
                ],
              ),
      ),
    );
  }
}
