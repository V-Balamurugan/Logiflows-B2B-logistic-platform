import 'package:flutter_test/flutter_test.dart';
import 'package:logiflows_mobile/main.dart';

void main() {
  testWidgets('Employee app smoke and foundation render test', (WidgetTester tester) async {
    await tester.pumpWidget(const LogiFlowsEmployeeApp());

    // Verify title and foundation card render
    expect(find.text('LogiFlows Courier'), findsOneWidget);
    expect(find.text('Phase 0 Foundation Ready'), findsOneWidget);
  });
}
