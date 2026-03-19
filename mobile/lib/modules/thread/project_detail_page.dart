import 'package:flutter/material.dart';

import '../../app/app_routes.dart';
import '../../app/app_scope.dart';
import '../../app/display_text.dart';
import '../../services/real/server_error_presenter.dart';
import '../message/thread_detail_page.dart';
import '../file/file_browser_page.dart';
import '../project/project_record.dart';
import 'thread_record.dart';
import 'thread_list_controller.dart';

class ProjectDetailPage extends StatefulWidget {
  const ProjectDetailPage({
    required this.project,
    super.key,
  });

  final ProjectRecord project;

  @override
  State<ProjectDetailPage> createState() => _ProjectDetailPageState();
}

class _ProjectDetailPageState extends State<ProjectDetailPage> {
  ThreadListController? _controller;
  String? _currentProjectId;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    final services = AppScope.servicesOf(context);
    if (_controller != null && _currentProjectId == widget.project.id) {
      return;
    }

    _currentProjectId = widget.project.id;
    _controller?.dispose();
    _controller = ThreadListController(
      services.restClient,
      services.webSocketClient,
    )
      ..loadThreads(widget.project.id)
      ..startRealtime(widget.project.id);
  }

  @override
  void dispose() {
    _controller?.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final controller = _controller;
    if (controller == null) {
      return const SizedBox.shrink();
    }

    return AnimatedBuilder(
      animation: controller,
      builder: (BuildContext context, Widget? child) {
        return Scaffold(
          appBar: AppBar(
            title: Text('${widget.project.name} Workspace'),
            actions: <Widget>[
              IconButton(
                tooltip: 'Create thread',
                onPressed: () => _createThread(context, controller),
                icon: const Icon(Icons.add_comment_rounded),
              ),
              IconButton(
                tooltip: 'Browse project files',
                onPressed: () {
                  Navigator.of(context).push(
                    MaterialPageRoute<void>(
                      builder: (_) =>
                          FileBrowserPage(projectId: widget.project.id),
                    ),
                  );
                },
                icon: const Icon(Icons.folder_open_rounded),
              ),
            ],
          ),
          body: Column(
            children: <Widget>[
              Padding(
                padding: const EdgeInsets.fromLTRB(16, 12, 16, 8),
                child: Card(
                  child: Padding(
                    padding: const EdgeInsets.all(16),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: <Widget>[
                        Text(
                          widget.project.workspaceRoot,
                          style: Theme.of(context).textTheme.titleSmall,
                        ),
                        const SizedBox(height: 12),
                        Wrap(
                          spacing: 8,
                          runSpacing: 8,
                          children: <Widget>[
                            _ProjectMetaChip(
                              label:
                                  'Created ${formatTimestamp(widget.project.createdAt)}',
                            ),
                            _ProjectMetaChip(
                              label:
                                  'Active ${formatTimestamp(widget.project.lastActivityAt)}',
                            ),
                            _ProjectMetaChip(
                              label: '${widget.project.threadCount} threads',
                            ),
                            _ProjectMetaChip(label: widget.project.machineId),
                          ],
                        ),
                      ],
                    ),
                  ),
                ),
              ),
              Expanded(
                child: controller.isLoading && controller.threads.isEmpty
                    ? const Center(child: CircularProgressIndicator())
                    : controller.errorMessage != null &&
                            controller.threads.isEmpty
                        ? Center(child: Text(controller.errorMessage!))
                        : ListView.builder(
                            padding: const EdgeInsets.fromLTRB(16, 8, 16, 24),
                            itemCount: controller.threads.length,
                            itemBuilder: (BuildContext context, int index) {
                              final thread = controller.threads[index];
                              return Padding(
                                padding: const EdgeInsets.only(bottom: 14),
                                child: ListTile(
                                  shape: RoundedRectangleBorder(
                                    borderRadius: BorderRadius.circular(20),
                                  ),
                                  tileColor:
                                      Theme.of(context).colorScheme.surface,
                                  contentPadding: const EdgeInsets.symmetric(
                                    horizontal: 18,
                                    vertical: 14,
                                  ),
                                  minVerticalPadding: 14,
                                  onTap: () {
                                    Navigator.of(context).pushNamed(
                                      AppRoutes.threadDetail,
                                      arguments: thread,
                                    );
                                  },
                                  title: Text(thread.title),
                                  subtitle: Padding(
                                    padding: const EdgeInsets.only(top: 8),
                                    child: Column(
                                      crossAxisAlignment:
                                          CrossAxisAlignment.start,
                                      children: <Widget>[
                                        Text(
                                          thread.displaySummary,
                                          maxLines: 2,
                                          overflow: TextOverflow.ellipsis,
                                        ),
                                        const SizedBox(height: 8),
                                        Wrap(
                                          spacing: 8,
                                          runSpacing: 8,
                                          children: <Widget>[
                                            _ProjectMetaChip(
                                              label: thread.agentLabel,
                                            ),
                                            _ProjectMetaChip(
                                              label:
                                                  'Started ${formatTimestamp(thread.startedAt)}',
                                            ),
                                            _ProjectMetaChip(
                                              label:
                                                  'Active ${formatTimestamp(thread.lastActivityAt)}',
                                            ),
                                          ],
                                        ),
                                      ],
                                    ),
                                  ),
                                  trailing: Text(thread.status.label),
                                ),
                              );
                            },
                          ),
              ),
            ],
          ),
        );
      },
    );
  }

  Future<void> _createThread(
    BuildContext context,
    ThreadListController controller,
  ) async {
    final services = AppScope.servicesOf(context);
    final prompt = await _showCreateThreadDialog(context);
    if (!mounted || prompt == null) {
      return;
    }

    try {
      final thread =
          await services.restClient.createProjectThread(widget.project.id, prompt);
      if (!mounted) {
        return;
      }
      controller.prependThread(thread);
      await Navigator.of(context).push(
        MaterialPageRoute<void>(
          builder: (_) => ThreadDetailPage(thread: thread),
        ),
      );
      if (mounted) {
        await controller.loadThreads(widget.project.id);
      }
    } catch (error) {
      if (!mounted) {
        return;
      }
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(presentServerError(error))),
      );
    }
  }

  Future<String?> _showCreateThreadDialog(BuildContext context) async {
    final controller = TextEditingController();
    try {
      return await showDialog<String>(
        context: context,
        builder: (BuildContext context) {
          return AlertDialog(
            title: const Text('Create Thread'),
            content: TextField(
              controller: controller,
              autofocus: true,
              minLines: 3,
              maxLines: 6,
              decoration: const InputDecoration(
                hintText: 'Enter the first prompt for this project thread',
              ),
            ),
            actions: <Widget>[
              TextButton(
                onPressed: () => Navigator.of(context).pop(),
                child: const Text('Cancel'),
              ),
              FilledButton(
                onPressed: () {
                  final value = controller.text.trim();
                  if (value.isEmpty) {
                    return;
                  }
                  Navigator.of(context).pop(value);
                },
                child: const Text('Create'),
              ),
            ],
          );
        },
      );
    } finally {
      controller.dispose();
    }
  }
}

class _ProjectMetaChip extends StatelessWidget {
  const _ProjectMetaChip({required this.label});

  final String label;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
      decoration: BoxDecoration(
        color: Theme.of(context).colorScheme.surfaceContainerHighest,
        borderRadius: BorderRadius.circular(999),
      ),
      child: Text(label, style: Theme.of(context).textTheme.bodySmall),
    );
  }
}
