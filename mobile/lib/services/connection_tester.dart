import '../app/app_environment.dart';

enum ConnectionTestStatus {
  success,
  failure,
  skipped,
}

class ConnectionTestResult {
  const ConnectionTestResult({
    required this.status,
    required this.message,
  });

  final ConnectionTestStatus status;
  final String message;

  bool get isSuccess => status == ConnectionTestStatus.success;
}

abstract class CodeScopeConnectionTester {
  Future<ConnectionTestResult> test(AppEnvironment environment);
}
