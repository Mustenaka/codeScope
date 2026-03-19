import 'package:flutter/material.dart';
import '../../l10n/generated/app_localizations.dart';

import '../../app/app_routes.dart';

class SettingsButton extends StatelessWidget {
  const SettingsButton({super.key});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return IconButton(
      tooltip: l10n.settingsTooltip,
      onPressed: () {
        Navigator.of(context).pushNamed(AppRoutes.settings);
      },
      icon: const Icon(Icons.settings_rounded),
    );
  }
}
