import 'package:flutter/material.dart';

import 'app/app_environment.dart';
import 'app/app_shell.dart';
import 'app/app_version.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await AppVersion.initialize();
  final environment = AppEnvironment.fromDartDefines();
  runApp(CodeScopeApp(initialEnvironment: environment));
}
