import 'package:flutter/material.dart';

import '../../app/app_routes.dart';
import '../../app/app_scope.dart';
import '../../services/app_services.dart';
import '../settings/settings_button.dart';
import 'project_controller.dart';
import 'project_record.dart';

class ProjectPage extends StatefulWidget {
  const ProjectPage({super.key});

  @override
  State<ProjectPage> createState() => _ProjectPageState();
}

class _ProjectPageState extends State<ProjectPage> {
  ProjectController? _controller;
  AppServices? _services;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    final services = AppScope.servicesOf(context);
    if (_services == services) {
      return;
    }
    _services = services;
    _controller?.dispose();
    _controller = ProjectController(
      services.restClient,
      services.webSocketClient,
    );
    _controller!
      ..loadProjects()
      ..startRealtime();
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
            title: const Text('Projects'),
            actions: const <Widget>[SettingsButton()],
          ),
          body: RefreshIndicator(
            onRefresh: controller.loadProjects,
            child: ListView(
              physics: const AlwaysScrollableScrollPhysics(),
              padding: const EdgeInsets.fromLTRB(16, 12, 16, 24),
              children: <Widget>[
                _HeaderBanner(useMock: _services!.environment.useMock),
                const SizedBox(height: 16),
                if (controller.isLoading && controller.projects.isEmpty)
                  const Padding(
                    padding: EdgeInsets.only(top: 80),
                    child: Center(child: CircularProgressIndicator()),
                  )
                else if (controller.errorMessage != null &&
                    controller.projects.isEmpty)
                  _ErrorState(
                    message: controller.errorMessage!,
                    onRetry: controller.loadProjects,
                  )
                else
                  ...controller.projects.map(
                    (ProjectRecord project) => Padding(
                      padding: const EdgeInsets.only(bottom: 12),
                      child: Card(
                        child: ListTile(
                          onTap: () {
                            Navigator.of(context).pushNamed(
                              AppRoutes.projectDetail,
                              arguments: project,
                            );
                          },
                          title: Text(project.name),
                          subtitle: Text(project.workspaceRoot),
                          trailing: Column(
                            mainAxisAlignment: MainAxisAlignment.center,
                            crossAxisAlignment: CrossAxisAlignment.end,
                            children: <Widget>[
                              Text('${project.threadCount} threads'),
                              Text('${project.runningThreadCount} running'),
                            ],
                          ),
                        ),
                      ),
                    ),
                  ),
              ],
            ),
          ),
        );
      },
    );
  }
}

class _HeaderBanner extends StatelessWidget {
  const _HeaderBanner({required this.useMock});

  final bool useMock;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Container(
      padding: const EdgeInsets.all(18),
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(24),
        gradient: const LinearGradient(
          colors: <Color>[Color(0xFF0F766E), Color(0xFF164E63)],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Text(
            'Projects',
            style: theme.textTheme.headlineSmall?.copyWith(
              color: Colors.white,
              fontWeight: FontWeight.w700,
            ),
          ),
          const SizedBox(height: 10),
          Text(
            useMock
                ? 'Project -> thread -> message view running on mock data.'
                : 'Project -> thread -> message view connected to server.',
            style: theme.textTheme.bodyMedium?.copyWith(
              color: Colors.white.withValues(alpha: 0.84),
            ),
          ),
        ],
      ),
    );
  }
}

class _ErrorState extends StatelessWidget {
  const _ErrorState({
    required this.message,
    required this.onRetry,
  });

  final String message;
  final Future<void> Function() onRetry;

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          children: <Widget>[
            const Icon(Icons.error_outline_rounded, size: 36),
            const SizedBox(height: 12),
            Text(message, textAlign: TextAlign.center),
            const SizedBox(height: 12),
            FilledButton(
              onPressed: onRetry,
              child: const Text('Retry'),
            ),
          ],
        ),
      ),
    );
  }
}
