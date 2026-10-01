import 'package:flutter_test/flutter_test.dart';

import 'package:parent_app/main.dart';

void main() {
  testWidgets('starts on the family workspace', (tester) async {
    await tester.pumpWidget(const ParentApp());
    await tester.pumpAndSettle();

    expect(find.text('如此萌屋'), findsOneWidget);
    expect(find.text('Family workspace'), findsOneWidget);
    expect(find.text('View devices'), findsOneWidget);
  });
}
