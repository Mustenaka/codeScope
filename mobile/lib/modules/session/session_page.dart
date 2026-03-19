import 'package:flutter/material.dart';
import '../../l10n/generated/app_localizations.dart';

import '../../app/app_routes.dart';
import '../../app/app_scope.dart';
import '../../services/app_services.dart';
import 'project_group.dart';
import '../settings/settings_button.dart';
import 'session_card.dart';
import 'session_controller.dart';
import 'session_record.dart';

class SessionPage extends StatefulWidget {
  const SessionPage({super.key});

  @override
  State<SessionPage> createState() => _SessionPageState();
}

class _SessionPageState extends State<SessionPage> {
  SessionController? _controller;
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
    _controller = SessionController(services.restClient);
    _controller!.loadSessions();
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
        final l10n = AppLocalizations.of(context)!;
        return Scaffold(
          appBar: AppBar(
            title: Text(l10n.sessionsTitle),
            actions: const <Widget>[SettingsButton()],
          ),
          body: RefreshIndicator(
            onRefresh: controller.loadSessions,
            child: ListView(
              physics: const AlwaysScrollableScrollPhysics(),
              padding: const EdgeInsets.fromLTRB(16, 12, 16, 24),
              children: <Widget>[
                _HeaderBanner(
                  useMock: _services!.environment.useMock,
                  title: l10n.remoteSessionsHeadline,
                  summaryText: l10n.sessionsReady(controller.sessions.length),
                  mockSuffix: l10n.mockTransportSuffix,
                ),
                const SizedBox(height: 16),
                if (controller.isLoading && controller.sessions.isEmpty)
                  const Padding(
                    padding: EdgeInsets.only(top: 80),
                    child: Center(child: CircularProgressIndicator()),
                  )
                else if (controller.errorMessage != null &&
                    controller.sessions.isEmpty)
                  _ErrorState(
                    message: l10n.loadingErrorPrefix(controller.errorMessage!),
                    onRetry: controller.loadSessions,
                  )
                else
                  ...controller.projectGroups.map(
                    (ProjectGroup group) => Padding(
                      padding: const EdgeInsets.only(bottom: 12),
                      child: _ProjectGroupCard(group: group),
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

class _ProjectGroupCard extends StatelessWidget {
  const _ProjectGroupCard({required this.group});

  final ProjectGroup group;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Card(
      child: ExpansionTile(
        tilePadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
        childrenPadding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
        title: Text(
          group.projectName,
          style: theme.textTheme.titleMedium?.copyWith(
            fontWeight: FontWeight.w700,
          ),
        ),
        subtitle: Padding(
          padding: const EdgeInsets.only(top: 6),
          child: Text(
            group.workspaceRoot,
            style: theme.textTheme.bodySmall?.copyWith(
              color: theme.colorScheme.onSurfaceVariant,
            ),
          ),
        ),
        children: <Widget>[
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: <Widget>[
              _ProjectMetaPill(label: '${group.sessions.length} agents'),
              _ProjectMetaPill(label: '${group.runningCount} running'),
              _ProjectMetaPill(
                label: '${group.waitingPromptCount} waiting prompt',
              ),
              _ProjectMetaPill(
                label: '${group.waitingReviewCount} waiting review',
              ),
              _ProjectMetaPill(label: _formatTime(group.updatedAt)),
            ],
          ),
          const SizedBox(height: 14),
          ...group.sessions.map(
            (SessionRecord session) => Padding(
              padding: const EdgeInsets.only(bottom: 10),
              child: SessionCard(
                session: session,
                onTap: () {
                  Navigator.of(context).pushNamed(
                    AppRoutes.sessionDetail,
                    arguments: session.id,
                  );
                },
              ),
            ),
          ),
        ],
      ),
    );
  }

  String _formatTime(DateTime dateTime) {
    final local = dateTime.toLocal();
    final month = local.month.toString().padLeft(2, '0');
    final day = local.day.toString().padLeft(2, '0');
    final hour = local.hour.toString().padLeft(2, '0');
    final minute = local.minute.toString().padLeft(2, '0');
    return '$month/$day $hour:$minute';
  }
}

class _ProjectMetaPill extends StatelessWidget {
  const _ProjectMetaPill({required this.label});

  final String label;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return DecoratedBox(
      decoration: BoxDecoration(
        color: theme.colorScheme.surfaceContainerHighest,
        borderRadius: BorderRadius.circular(999),
      ),
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
        child: Text(label, style: theme.textTheme.labelMedium),
      ),
    );
  }
}

class _HeaderBanner extends StatelessWidget {
  const _HeaderBanner({
    required this.useMock,
    required this.title,
    required this.summaryText,
    required this.mockSuffix,
  });

  final bool useMock;
  final String title;
  final String summaryText;
  final String mockSuffix;

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
            title,
            style: theme.textTheme.headlineSmall?.copyWith(
              color: Colors.white,
              fontWeight: FontWeight.w700,
            ),
          ),
          const SizedBox(height: 10),
          Text(
            useMock ? '$summaryText - $mockSuffix' : summaryText,
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
    final l10n = AppLocalizations.of(context)!;
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          children: <Widget>[
            const Icon(Icons.error_outline_rounded, size: 36),
            const SizedBox(height: 12),
            Text(
              message,
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 12),
            FilledButton(
              onPressed: onRetry,
              child: Text(l10n.retry),
            ),
          ],
        ),
      ),
    );
  }
}
