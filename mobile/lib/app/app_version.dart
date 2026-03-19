import 'package:package_info_plus/package_info_plus.dart';

/// Holds the app version read from pubspec.yaml at startup.
///
/// Call [AppVersion.initialize] once in [main] before [runApp].
/// After that, [AppVersion.instance] is always available synchronously.
class AppVersion {
  const AppVersion._({
    required this.version,
    required this.buildNumber,
  });

  /// Semantic version string, e.g. "0.1.0".
  final String version;

  /// Build number string, e.g. "1".
  final String buildNumber;

  /// Full display string, e.g. "0.1.0+1".
  String get display => '$version+$buildNumber';

  static const AppVersion _fallback = AppVersion._(
    version: '0.0.0',
    buildNumber: '0',
  );

  static AppVersion? _instance;

  static AppVersion get instance => _instance ?? _fallback;

  static Future<void> initialize() async {
    final info = await PackageInfo.fromPlatform();
    _instance = AppVersion._(
      version: info.version,
      buildNumber: info.buildNumber,
    );
  }
}
