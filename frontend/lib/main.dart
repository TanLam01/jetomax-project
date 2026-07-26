import 'package:flutter/widgets.dart';

import 'app.dart';
import 'core/dependencies.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  final dependencies = await AppDependencies.create();
  runApp(ChatApp(dependencies: dependencies));
}
