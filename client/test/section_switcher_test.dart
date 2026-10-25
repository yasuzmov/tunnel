import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:tunnel_client/app/theme.dart';
import 'package:tunnel_client/shell/section_switcher.dart';

void main() {
  testWidgets('section switcher animates between sections', (tester) async {
    await tester.pumpWidget(MaterialApp(
      theme: buildAppTheme(),
      home: Scaffold(
        body: SectionSwitcher(
          index: 0,
          children: const [Center(child: Text('A')), Center(child: Text('B'))],
        ),
      ),
    ));
    await tester.pumpAndSettle();
    expect(find.text('A'), findsOneWidget);

    await tester.pumpWidget(MaterialApp(
      theme: buildAppTheme(),
      home: Scaffold(
        body: SectionSwitcher(
          index: 1,
          children: const [Center(child: Text('A')), Center(child: Text('B'))],
        ),
      ),
    ));
    await tester.pump(const Duration(milliseconds: 120));
    expect(tester.takeException(), isNull);
    await tester.pumpAndSettle();
    expect(find.text('B'), findsOneWidget);
    expect(tester.takeException(), isNull);
  });
}
