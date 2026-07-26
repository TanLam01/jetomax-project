import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:get_it/get_it.dart';

import '../data/chat_repository.dart';
import '../realtime/realtime_client.dart';
import 'api_client.dart';
import 'token_store.dart';

class AppDependencies {
  AppDependencies._(this.getIt);
  final GetIt getIt;

  static Future<AppDependencies> create() async {
    final getIt = GetIt.asNewInstance();
    const secureStorage = FlutterSecureStorage();
    final tokens = TokenStore(secureStorage);
    await tokens.load();
    final api = ApiClient(tokens);
    getIt
      ..registerSingleton<TokenStore>(tokens)
      ..registerSingleton<ApiClient>(api)
      ..registerLazySingleton(() => ChatRepository(api))
      ..registerLazySingleton(() => RealtimeClient(tokens));
    return AppDependencies._(getIt);
  }
}
