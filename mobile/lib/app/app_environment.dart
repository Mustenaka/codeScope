class AppEnvironment {
  const AppEnvironment({
    required this.useMock,
    required this.apiBaseUrl,
    required this.webSocketUrl,
  });

  static const String defaultApiBaseUrl = 'http://192.168.31.137:8080/api';
  static const String defaultWebSocketUrl =
      'ws://192.168.31.137:8080/ws/mobile';

  final bool useMock;
  final String apiBaseUrl;
  final String webSocketUrl;

  static const AppEnvironment mock = AppEnvironment(
    useMock: true,
    apiBaseUrl: defaultApiBaseUrl,
    webSocketUrl: defaultWebSocketUrl,
  );

  factory AppEnvironment.fromDartDefines() {
    return AppEnvironment.fromValues(
      useMockValue: const String.fromEnvironment(
        'CODESCOPE_USE_MOCK',
        defaultValue: '',
      ),
      apiBaseUrlValue: const String.fromEnvironment(
        'CODESCOPE_API_BASE_URL',
        defaultValue: '',
      ),
      webSocketUrlValue: const String.fromEnvironment(
        'CODESCOPE_WS_URL',
        defaultValue: '',
      ),
    );
  }

  factory AppEnvironment.fromValues({
    String? useMockValue,
    String? apiBaseUrlValue,
    String? webSocketUrlValue,
  }) {
    final normalizedApiBaseUrl =
        _normalizeValue(apiBaseUrlValue) ?? defaultApiBaseUrl;
    final normalizedWebSocketUrl =
        _normalizeValue(webSocketUrlValue) ?? defaultWebSocketUrl;
    final parsedUseMock = _parseBool(useMockValue);
    final hasServerOverrides = _normalizeValue(apiBaseUrlValue) != null ||
        _normalizeValue(webSocketUrlValue) != null;

    return AppEnvironment(
      useMock: parsedUseMock ?? !hasServerOverrides,
      apiBaseUrl: normalizedApiBaseUrl,
      webSocketUrl: normalizedWebSocketUrl,
    );
  }

  static String? _normalizeValue(String? value) {
    final trimmed = value?.trim();
    if (trimmed == null || trimmed.isEmpty) {
      return null;
    }
    return trimmed;
  }

  static bool? _parseBool(String? value) {
    switch (_normalizeValue(value)?.toLowerCase()) {
      case 'true':
        return true;
      case 'false':
        return false;
    }
    return null;
  }
}
