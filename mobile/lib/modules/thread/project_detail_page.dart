import 'package:flutter/material.dart';

import '../../app/app_routes.dart';
import '../../app/app_scope.dart';
import '../project/project_record.dart';
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
          appBar: AppBar(title: Text(widget.project.name)),
          body: Column(
            children: <Widget>[
              ListTile(
                title: Text(widget.project.workspaceRoot),
                subtitle: Text('${widget.project.threadCount} threads'),
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
                                    child: Text(
                                      thread.displaySummary,
                                      maxLines: 2,
                                      overflow: TextOverflow.ellipsis,
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
}
