import 'package:flutter/material.dart';
import '../../l10n/generated/app_localizations.dart';

import '../../app/app_localization_x.dart';
import 'log_event.dart';
import 'log_event_view_model.dart';

class LogEventTile extends StatelessWidget {
  const LogEventTile({
    required this.event,
    super.key,
  });

  final LogEvent event;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final l10n = AppLocalizations.of(context)!;
    final scheme = theme.colorScheme;
    final viewModel = LogEventViewModel(event);
    final color = switch (event.type) {
      LogEventType.aiOutput => const Color(0xFF0F766E),
      LogEventType.terminalOutput => const Color(0xFF1D4ED8),
      LogEventType.command => const Color(0xFF7C3AED),
      LogEventType.commandResult => const Color(0xFF4338CA),
      LogEventType.fileChange => const Color(0xFFB45309),
      LogEventType.diff => const Color(0xFF047857),
      LogEventType.error => const Color(0xFFB91C1C),
      LogEventType.heartbeat => scheme.onSurfaceVariant,
    };

    return Container(
      margin: const EdgeInsets.only(bottom: 10),
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(
          color: scheme.outlineVariant.withValues(alpha: 0.45),
        ),
      ),
      child: Theme(
        data: theme.copyWith(dividerColor: Colors.transparent),
        child: ExpansionTile(
          tilePadding: EdgeInsets.zero,
          childrenPadding: EdgeInsets.zero,
          title: Row(
            children: <Widget>[
              Container(
                padding:
                    const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                decoration: BoxDecoration(
                  color: color.withValues(alpha: 0.12),
                  borderRadius: BorderRadius.circular(999),
                ),
                child: Text(
                  event.type.localized(l10n),
                  style: theme.textTheme.labelMedium?.copyWith(
                    color: color,
                    fontWeight: FontWeight.w700,
                  ),
                ),
              ),
              const Spacer(),
              Text(
                _formatTime(event.createdAt),
                style: theme.textTheme.labelMedium?.copyWith(
                  color: scheme.onSurfaceVariant,
                ),
              ),
            ],
          ),
          subtitle: Padding(
            padding: const EdgeInsets.only(top: 10),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                Text(
                  viewModel.summary,
                  style: theme.textTheme.bodyMedium?.copyWith(height: 1.45),
                ),
                if (viewModel.highlights.isNotEmpty) ...<Widget>[
                  const SizedBox(height: 10),
                  Wrap(
                    spacing: 8,
                    runSpacing: 8,
                    children: viewModel.highlights
                        .map((MapEntry<String, Object?> entry) {
                      return DecoratedBox(
                        decoration: BoxDecoration(
                          color: scheme.surfaceContainerHighest,
                          borderRadius: BorderRadius.circular(999),
                        ),
                        child: Padding(
                          padding: const EdgeInsets.symmetric(
                              horizontal: 10, vertical: 5),
                          child: Text(
                            '${entry.key}: ${entry.value}',
                            style: theme.textTheme.labelSmall,
                          ),
                        ),
                      );
                    }).toList(),
                  ),
                ],
              ],
            ),
          ),
          children: <Widget>[
            if (viewModel.rawMetadata.isNotEmpty) ...<Widget>[
              const SizedBox(height: 10),
              Align(
                alignment: Alignment.centerLeft,
                child: Text(
                  'Raw payload',
                  style: theme.textTheme.labelMedium?.copyWith(
                    fontWeight: FontWeight.w700,
                    color: scheme.onSurfaceVariant,
                  ),
                ),
              ),
              const SizedBox(height: 8),
              ...viewModel.rawMetadata.map(
                (MapEntry<String, Object?> entry) => Padding(
                  padding: const EdgeInsets.only(bottom: 6),
                  child: Row(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: <Widget>[
                      SizedBox(
                        width: 110,
                        child: Text(
                          entry.key,
                          style: theme.textTheme.labelMedium?.copyWith(
                            color: scheme.onSurfaceVariant,
                          ),
                        ),
                      ),
                      Expanded(
                        child: Text(
                          '${entry.value}',
                          style: theme.textTheme.bodySmall,
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }

  String _formatTime(DateTime dateTime) {
    final local = dateTime.toLocal();
    final hour = local.hour.toString().padLeft(2, '0');
    final minute = local.minute.toString().padLeft(2, '0');
    final second = local.second.toString().padLeft(2, '0');
    return '$hour:$minute:$second';
  }
}
