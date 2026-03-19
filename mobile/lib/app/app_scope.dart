import 'package:flutter/widgets.dart';

import '../services/app_services.dart';
import 'app_settings.dart';

class AppScope extends InheritedWidget {
  const AppScope({
    required this.services,
    required this.settingsController,
    required super.child,
    super.key,
  });

  final AppServices services;
  final AppSettingsController settingsController;

  static AppServices servicesOf(BuildContext context) {
    final scope = context.dependOnInheritedWidgetOfExactType<AppScope>();
    assert(scope != null, 'AppScope not found in widget tree.');
    return scope!.services;
  }

  static AppSettingsController settingsOf(BuildContext context) {
    final scope = context.dependOnInheritedWidgetOfExactType<AppScope>();
    assert(scope != null, 'AppScope not found in widget tree.');
    return scope!.settingsController;
  }

  @override
  bool updateShouldNotify(AppScope oldWidget) {
    return oldWidget.services != services ||
        oldWidget.settingsController != settingsController;
  }
}
