import 'package:flutter/material.dart';

import 'app/app.dart';

export 'app/app.dart' show ParentApp;

void main() {
  WidgetsFlutterBinding.ensureInitialized();
  runApp(const ParentApp());
}
