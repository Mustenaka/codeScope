import 'package:flutter/material.dart';
import '../../l10n/generated/app_localizations.dart';

import '../../app/app_localization_x.dart';
import '../../app/app_scope.dart';
import '../file/file_browser_page.dart';
import '../prompt/prompt_page.dart';
import '../session/session_record.dart';
import '../settings/settings_button.dart';
import 'log_controller.dart';
import 'log_event_tile.dart';

class LogPage extends StatefulWidget {
  const LogPage({
    required this.sessionId,
    super.key,
  });

  final String sessionId;

  @override
  State<LogPage> createState() => _LogPageState();
}

class _LogPageState extends State<LogPage> {
  LogController? _controller;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    _controller ??= LogController(
      restClient: AppScope.servicesOf(context).restClient,
      webSocketClient: AppScope.servicesOf(context).webSocketClient,
    )..load(widget.sessionId);
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
        final session = controller.session;
        final digest = controller.digest;

        return Scaffold(
          appBar: AppBar(
            title: Text(session?.projectName ?? l10n.sessionDetailsTitle),
            actions: const <Widget>[SettingsButton()],
          ),
          body: Column(
            children: <Widget>[
              if (session != null)
                _SessionSummary(
                  projectName: session.projectName,
                  machineId: session.machineId,
                  workspaceRoot: session.workspaceRoot,
                  status: session.status.localized(l10n),
                  summary: digest?.summary ?? session.summary,
                  agentState: session.agentState,
                ),
              Expanded(
                child: controller.isLoading &&
                        (digest?.visibleEvents.isEmpty ?? true)
                    ? const Center(child: CircularProgressIndicator())
                    : controller.errorMessage != null &&
                            (digest?.visibleEvents.isEmpty ?? true)
                        ? Center(
                            child: Text(
                              l10n.loadingErrorPrefix(controller.errorMessage!),
                            ),
                          )
                        : ListView.builder(
                            padding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
                            itemCount: (digest?.visibleEvents.length ?? 0) + 1,
                            itemBuilder: (BuildContext context, int index) {
                              if (index == 0) {
                                return _ActionStrip(
                                    sessionId: widget.sessionId);
                              }
                              return LogEventTile(
                                event: digest!.visibleEvents[index - 1],
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

class _SessionSummary extends StatelessWidget {
  const _SessionSummary({
    required this.projectName,
    required this.machineId,
    required this.workspaceRoot,
    required this.status,
    required this.summary,
    required this.agentState,
  });

  final String projectName;
  final String machineId;
  final String workspaceRoot;
  final String status;
  final String summary;
  final AgentSummaryState agentState;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final theme = Theme.of(context);
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 12, 16, 12),
      child: Card(
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: <Widget>[
              Text(
                projectName,
                style: theme.textTheme.titleLarge?.copyWith(
                  fontWeight: FontWeight.w700,
                ),
              ),
              const SizedBox(height: 10),
              Wrap(
                spacing: 8,
                runSpacing: 8,
                children: <Widget>[
                  _InfoChip(
                    icon: Icons.smart_toy_rounded,
                    label: 'Agent: ${_agentStateLabel(agentState)}',
                  ),
                  _InfoChip(
                    icon: Icons.memory_rounded,
                    label: '${l10n.machineLabel}: $machineId',
                  ),
                  _InfoChip(
                    icon: Icons.folder_rounded,
                    label: '${l10n.workspaceLabel}: $workspaceRoot',
                  ),
                  _InfoChip(
                    icon: Icons.circle_rounded,
                    label: '${l10n.statusLabel}: $status',
                  ),
                ],
              ),
              if (summary.isNotEmpty) ...<Widget>[
                const SizedBox(height: 12),
                Text(
                  summary,
                  style: theme.textTheme.bodyMedium?.copyWith(
                    color: theme.colorScheme.onSurfaceVariant,
                    height: 1.4,
                  ),
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }

  String _agentStateLabel(AgentSummaryState state) {
    return switch (state) {
      AgentSummaryState.running => 'Running',
      AgentSummaryState.waitingPrompt => 'Waiting Prompt',
      AgentSummaryState.waitingReview => 'Waiting Review',
      AgentSummaryState.completed => 'Completed',
      AgentSummaryState.blocked => 'Blocked',
    };
  }
}

class _InfoChip extends StatelessWidget {
  const _InfoChip({
    required this.icon,
    required this.label,
  });

  final IconData icon;
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
        padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 7),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: <Widget>[
            Icon(icon, size: 14),
            const SizedBox(width: 6),
            Text(label, style: theme.textTheme.labelMedium),
          ],
        ),
      ),
    );
  }
}

class _ActionStrip extends StatelessWidget {
  const _ActionStrip({required this.sessionId});

  final String sessionId;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return Padding(
      padding: const EdgeInsets.only(bottom: 14),
      child: Row(
        children: <Widget>[
          Expanded(
            child: OutlinedButton.icon(
              onPressed: () {
                Navigator.of(context).push(
                  MaterialPageRoute<void>(
                    builder: (_) => FileBrowserPage(sessionId: sessionId),
                  ),
                );
              },
              icon: const Icon(Icons.folder_open_rounded),
              label: Text(l10n.files),
            ),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: OutlinedButton.icon(
              onPressed: () {
                Navigator.of(context).push(
                  MaterialPageRoute<void>(
                    builder: (_) => PromptPage(sessionId: sessionId),
                  ),
                );
              },
              icon: const Icon(Icons.bolt_rounded),
              label: Text(l10n.prompt),
            ),
          ),
        ],
      ),
    );
  }
}
