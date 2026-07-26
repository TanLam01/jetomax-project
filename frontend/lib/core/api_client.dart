import 'package:dio/dio.dart';

import '../models/models.dart';
import 'app_config.dart';
import 'token_store.dart';

class ApiClient {
  ApiClient(this.tokens)
    : dio = Dio(
        BaseOptions(
          baseUrl: AppConfig.apiBaseUrl,
          connectTimeout: const Duration(seconds: 10),
          receiveTimeout: const Duration(seconds: 15),
        ),
      ) {
    dio.interceptors.add(
      QueuedInterceptorsWrapper(
        onRequest: (options, handler) {
          final token = tokens.accessToken;
          if (token != null) options.headers['Authorization'] = 'Bearer $token';
          handler.next(options);
        },
        onError: (error, handler) async {
          final request = error.requestOptions;
          if (error.response?.statusCode != 401 ||
              request.path.contains('/auth/') ||
              request.extra['retried'] == true ||
              tokens.refreshToken == null) {
            handler.next(error);
            return;
          }
          try {
            final plain = Dio(BaseOptions(baseUrl: AppConfig.apiBaseUrl));
            final response = await plain.post<Map<String, dynamic>>(
              '/auth/refresh',
              data: {'refresh_token': tokens.refreshToken},
            );
            final session = Session.fromJson(response.data!);
            await tokens.save(session);
            request.extra['retried'] = true;
            request.headers['Authorization'] = 'Bearer ${session.accessToken}';
            handler.resolve(await dio.fetch(request));
          } catch (_) {
            await tokens.clear();
            onSessionExpired?.call();
            handler.next(error);
          }
        },
      ),
    );
  }

  final TokenStore tokens;
  final Dio dio;
  void Function()? onSessionExpired;
}

String readableApiError(Object error) {
  if (error is DioException) {
    final data = error.response?.data;
    if (data is Map && data['error'] is Map) {
      return (data['error']['message'] as String?) ?? 'Có lỗi xảy ra';
    }
    if (error.type == DioExceptionType.connectionError) {
      return 'Không thể kết nối tới máy chủ';
    }
  }
  return 'Có lỗi xảy ra, vui lòng thử lại';
}
