import 'package:flutter/material.dart';

import 'app_environment.dart';

enum AppLanguage {
  chinese,
  english;

  Locale get locale {
    switch (this) {
      case AppLanguage.chinese:
        return const Locale('zh');
      case AppLanguage.english:
        return const Locale('en');
    }
  }
}

class AppSettingsState {
  const AppSettingsState({
    required this.language,
    required this.useMock,
    required this.apiBaseUrl,
    required this.webSocketUrl,
  });

  final AppLanguage language;
  final bool useMock;
  final String apiBaseUrl;
  final String webSocketUrl;

  Locale get locale => language.locale;

  AppEnvironment toEnvironment() {
    return AppEnvironment(
      useMock: useMock,
      apiBaseUrl: apiBaseUrl,
      webSocketUrl: webSocketUrl,
    );
  }

  AppSettingsState copyWith({
    AppLanguage? language,
    bool? useMock,
    String? apiBaseUrl,
    String? webSocketUrl,
  }) {
    return AppSettingsState(
      language: language ?? this.language,
      useMock: useMock ?? this.useMock,
      apiBaseUrl: apiBaseUrl ?? this.apiBaseUrl,
      webSocketUrl: webSocketUrl ?? this.webSocketUrl,
    );
  }

  factory AppSettingsState.fromEnvironment(AppEnvironment environment) {
    return AppSettingsState(
      language: AppLanguage.chinese,
      useMock: environment.useMock,
      apiBaseUrl: environment.apiBaseUrl,
      webSocketUrl: environment.webSocketUrl,
    );
  }
}

class AppSettingsController extends ChangeNotifier {
  AppSettingsController(AppEnvironment initialEnvironment)
      : _state = AppSettingsState.fromEnvironment(initialEnvironment);

  AppSettingsState _state;

  AppSettingsState get state => _state;

  void apply(AppSettingsState nextState) {
    _state = nextState;
    notifyListeners();
  }
}
