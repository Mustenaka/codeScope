import 'package:codescope_mobile/app/app_environment.dart';
import 'package:codescope_mobile/app/app_shell.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  testWidgets('renders project list with mock content', (
    WidgetTester tester,
  ) async {
    await tester.pumpWidget(
      const CodeScopeApp(
        initialEnvironment: AppEnvironment.mock,
      ),
    );

    await tester.pumpAndSettle();

    expect(find.byIcon(Icons.settings_rounded), findsOneWidget);
    expect(find.text('codeScope'), findsOneWidget);
    expect(find.text('agent-bridge'), findsOneWidget);
  });

  testWidgets('navigates from project list to thread messages', (
    WidgetTester tester,
  ) async {
    await tester.pumpWidget(
      const CodeScopeApp(
        initialEnvironment: AppEnvironment.mock,
      ),
    );

    await tester.pumpAndSettle();

    await tester.tap(find.text('codeScope').first);
    await tester.pumpAndSettle();
    expect(find.text('Implemented server API'), findsWidgets);

    await tester.tap(find.text('Implemented server API').first);
    await tester.pumpAndSettle();
    expect(find.text('continue implementing'), findsWidgets);
  });
}
