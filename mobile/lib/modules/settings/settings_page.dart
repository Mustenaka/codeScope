import 'package:flutter/material.dart';
import '../../l10n/generated/app_localizations.dart';

import '../../app/app_environment.dart';
import '../../app/app_scope.dart';
import '../../app/app_settings.dart';
import '../../app/app_version.dart';
import '../../services/connection_tester.dart';

class SettingsPage extends StatefulWidget {
  const SettingsPage({super.key});

  @override
  State<SettingsPage> createState() => _SettingsPageState();
}

class _SettingsPageState extends State<SettingsPage> {
  AppSettingsController? _settingsController;
  late final TextEditingController _apiController;
  late final TextEditingController _webSocketController;
  late AppLanguage _language;
  late bool _useMock;
  bool _testingConnection = false;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    final controller = AppScope.settingsOf(context);
    if (_settingsController == controller) {
      return;
    }

    _settingsController = controller;
    final state = controller.state;
    _language = state.language;
    _useMock = state.useMock;
    _apiController = TextEditingController(text: state.apiBaseUrl);
    _webSocketController = TextEditingController(text: state.webSocketUrl);
  }

  @override
  void dispose() {
    _apiController.dispose();
    _webSocketController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final theme = Theme.of(context);

    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.settingsTitle),
      ),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: <Widget>[
          _SectionCard(
            title: l10n.settingsSectionGeneral,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                Text(
                  l10n.languageLabel,
                  style: theme.textTheme.titleMedium?.copyWith(
                    fontWeight: FontWeight.w700,
                  ),
                ),
                const SizedBox(height: 12),
                SegmentedButton<AppLanguage>(
                  segments: <ButtonSegment<AppLanguage>>[
                    ButtonSegment<AppLanguage>(
                      value: AppLanguage.chinese,
                      label: Text(l10n.languageChinese),
                    ),
                    ButtonSegment<AppLanguage>(
                      value: AppLanguage.english,
                      label: Text(l10n.languageEnglish),
                    ),
                  ],
                  selected: <AppLanguage>{_language},
                  onSelectionChanged: (Set<AppLanguage> selection) {
                    setState(() {
                      _language = selection.first;
                    });
                  },
                ),
              ],
            ),
          ),
          const SizedBox(height: 16),
          _SectionCard(
            title: l10n.settingsSectionConnection,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                Text(
                  l10n.dataSourceLabel,
                  style: theme.textTheme.titleMedium?.copyWith(
                    fontWeight: FontWeight.w700,
                  ),
                ),
                const SizedBox(height: 12),
                SegmentedButton<bool>(
                  segments: <ButtonSegment<bool>>[
                    ButtonSegment<bool>(
                      value: true,
                      label: Text(l10n.dataSourceMock),
                    ),
                    ButtonSegment<bool>(
                      value: false,
                      label: Text(l10n.dataSourceServer),
                    ),
                  ],
                  selected: <bool>{_useMock},
                  onSelectionChanged: (Set<bool> selection) {
                    setState(() {
                      _useMock = selection.first;
                    });
                  },
                ),
                const SizedBox(height: 16),
                TextField(
                  controller: _apiController,
                  decoration: InputDecoration(
                    labelText: l10n.apiBaseUrlLabel,
                    border: const OutlineInputBorder(),
                  ),
                ),
                const SizedBox(height: 12),
                TextField(
                  controller: _webSocketController,
                  decoration: InputDecoration(
                    labelText: l10n.webSocketUrlLabel,
                    border: const OutlineInputBorder(),
                  ),
                ),
                const SizedBox(height: 12),
                Text(
                  l10n.settingsHintRuntimeOnly,
                  style: theme.textTheme.bodySmall?.copyWith(
                    color: theme.colorScheme.onSurfaceVariant,
                  ),
                ),
                const SizedBox(height: 16),
                SizedBox(
                  width: double.infinity,
                  child: OutlinedButton.icon(
                    key: const ValueKey('settings-test-connection'),
                    onPressed:
                        _useMock || _testingConnection ? null : _testConnection,
                    icon: _testingConnection
                        ? const SizedBox(
                            width: 16,
                            height: 16,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )
                        : const Icon(Icons.wifi_find_rounded),
                    label: Text(
                      _testingConnection
                          ? l10n.settingsTestingConnection
                          : l10n.settingsTestConnection,
                    ),
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: 20),
          _SectionCard(
            title: l10n.settingsSectionAbout,
            child: _AboutInfo(l10n: l10n),
          ),
          const SizedBox(height: 20),
          FilledButton.icon(
            onPressed: _save,
            icon: const Icon(Icons.check_rounded),
            label: Text(l10n.saveSettings),
          ),
        ],
      ),
    );
  }

  void _save() {
    final nextState = _settingsController!.state.copyWith(
      language: _language,
      useMock: _useMock,
      apiBaseUrl: _apiController.text.trim(),
      webSocketUrl: _webSocketController.text.trim(),
    );
    _settingsController!.apply(nextState);

    final l10n = AppLocalizations.of(context)!;
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(l10n.settingsSaved)),
    );
    Navigator.of(context).pop();
  }

  Future<void> _testConnection() async {
    final services = AppScope.servicesOf(context);
    final messenger = ScaffoldMessenger.of(context);
    final l10n = AppLocalizations.of(context)!;
    final environment = AppEnvironment(
      useMock: _useMock,
      apiBaseUrl: _apiController.text.trim(),
      webSocketUrl: _webSocketController.text.trim(),
    );

    setState(() {
      _testingConnection = true;
    });

    final result = await services.connectionTester.test(environment);
    if (!mounted) {
      return;
    }

    setState(() {
      _testingConnection = false;
    });

    final message = switch (result.status) {
      ConnectionTestStatus.success => l10n.settingsConnectionSuccess,
      ConnectionTestStatus.skipped => l10n.settingsConnectionMockOnly,
      ConnectionTestStatus.failure => l10n.settingsConnectionFailure(
          result.message,
        ),
    };

    messenger.showSnackBar(SnackBar(content: Text(message)));
  }
}

class _AboutInfo extends StatelessWidget {
  const _AboutInfo({required this.l10n});

  final AppLocalizations l10n;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final version = AppVersion.instance;
    final labelStyle = theme.textTheme.bodyMedium?.copyWith(
      color: theme.colorScheme.onSurfaceVariant,
    );
    final valueStyle = theme.textTheme.bodyMedium?.copyWith(
      fontWeight: FontWeight.w600,
    );
    return Column(
      children: <Widget>[
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: <Widget>[
            Text(l10n.appVersionLabel, style: labelStyle),
            Text(version.version, style: valueStyle),
          ],
        ),
        const SizedBox(height: 8),
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: <Widget>[
            Text(l10n.appBuildLabel, style: labelStyle),
            Text(version.buildNumber, style: valueStyle),
          ],
        ),
      ],
    );
  }
}

class _SectionCard extends StatelessWidget {
  const _SectionCard({
    required this.title,
    required this.child,
  });

  final String title;
  final Widget child;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            Text(
              title,
              style: theme.textTheme.titleLarge?.copyWith(
                fontWeight: FontWeight.w700,
              ),
            ),
            const SizedBox(height: 16),
            child,
          ],
        ),
      ),
    );
  }
}
