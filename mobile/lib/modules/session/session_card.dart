import 'package:flutter/material.dart';
import 'session_record.dart';

class SessionCard extends StatelessWidget {
  const SessionCard({
    required this.session,
    required this.onTap,
    super.key,
  });

  final SessionRecord session;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Card(
      child: InkWell(
        borderRadius: BorderRadius.circular(18),
        onTap: onTap,
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: <Widget>[
              Row(
                children: <Widget>[
                  Expanded(
                    child: Text(
                      session.projectName,
                      style: theme.textTheme.titleMedium?.copyWith(
                        fontWeight: FontWeight.w700,
                      ),
                    ),
                  ),
                  _StatusChip(state: session.agentState),
                ],
              ),
              if (session.summary.isNotEmpty) ...<Widget>[
                const SizedBox(height: 10),
                Text(
                  session.summary,
                  style: theme.textTheme.bodyMedium?.copyWith(
                    color: theme.colorScheme.onSurfaceVariant,
                  ),
                ),
                const SizedBox(height: 14),
              ] else
                const SizedBox(height: 10),
              Text(
                session.workspaceRoot,
                style: theme.textTheme.bodySmall?.copyWith(
                  color: theme.colorScheme.onSurfaceVariant,
                ),
              ),
              const SizedBox(height: 8),
              Wrap(
                spacing: 8,
                runSpacing: 8,
                children: <Widget>[
                  _MetaPill(
                    icon: Icons.memory_rounded,
                    label: session.machineId,
                  ),
                  _MetaPill(
                    icon: Icons.tag_rounded,
                    label: session.id,
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _StatusChip extends StatelessWidget {
  const _StatusChip({required this.state});

  final AgentSummaryState state;

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    final (background, foreground, label) = switch (state) {
      AgentSummaryState.running => (
          const Color(0xFFD1FAE5),
          const Color(0xFF065F46),
          'Running',
        ),
      AgentSummaryState.waitingPrompt => (
          const Color(0xFFFEF3C7),
          const Color(0xFF92400E),
          'Waiting Prompt',
        ),
      AgentSummaryState.waitingReview => (
          const Color(0xFFE0F2FE),
          const Color(0xFF075985),
          'Waiting Review',
        ),
      AgentSummaryState.completed => (
          scheme.surfaceContainerHighest,
          scheme.onSurfaceVariant,
          'Completed',
        ),
      AgentSummaryState.blocked => (
          const Color(0xFFFEE2E2),
          const Color(0xFF991B1B),
          'Blocked',
        ),
    };

    return Chip(
      label: Text(label),
      backgroundColor: background,
      labelStyle: TextStyle(
        color: foreground,
        fontWeight: FontWeight.w600,
      ),
      side: BorderSide.none,
      visualDensity: VisualDensity.compact,
    );
  }
}

class _MetaPill extends StatelessWidget {
  const _MetaPill({
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
        padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: <Widget>[
            Icon(icon, size: 14, color: theme.colorScheme.onSurfaceVariant),
            const SizedBox(width: 6),
            Text(label, style: theme.textTheme.labelMedium),
          ],
        ),
      ),
    );
  }
}
