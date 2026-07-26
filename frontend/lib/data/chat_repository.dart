import 'dart:typed_data';

import 'package:dio/dio.dart';

import '../core/api_client.dart';
import '../models/models.dart';

class ChatRepository {
  ChatRepository(this.api);
  final ApiClient api;

  Future<Session> login(String email, String password) async {
    final response = await api.dio.post<Map<String, dynamic>>(
      '/auth/login',
      data: {'email': email, 'password': password},
    );
    return Session.fromJson(response.data!);
  }

  Future<Session> register(String name, String email, String password) async {
    final response = await api.dio.post<Map<String, dynamic>>(
      '/auth/register',
      data: {'display_name': name, 'email': email, 'password': password},
    );
    return Session.fromJson(response.data!);
  }

  Future<void> logout() async {
    if (api.tokens.refreshToken != null) {
      await api.dio.post(
        '/auth/logout',
        data: {'refresh_token': api.tokens.refreshToken},
      );
    }
  }

  Future<List<Conversation>> conversations() async {
    final response = await api.dio.get<Map<String, dynamic>>('/conversations');
    return (response.data!['data'] as List)
        .map((item) => Conversation.fromJson(item as Map<String, dynamic>))
        .toList();
  }

  Future<List<User>> searchUsers(String query) async {
    final response = await api.dio.get<Map<String, dynamic>>(
      '/users',
      queryParameters: {'query': query},
    );
    return (response.data!['data'] as List)
        .map((item) => User.fromJson(item as Map<String, dynamic>))
        .toList();
  }

  Future<Conversation> createDirect(String userId) async {
    final response = await api.dio.post<Map<String, dynamic>>(
      '/conversations/direct',
      data: {'user_id': userId},
    );
    return Conversation.fromJson(response.data!);
  }

  Future<MessagePage> messages(
    String conversationId, {
    String? cursor,
    String? after,
    int limit = 30,
  }) async {
    final response = await api.dio.get<Map<String, dynamic>>(
      '/conversations/$conversationId/messages',
      queryParameters: {
        if (cursor != null) 'cursor': cursor,
        if (after != null) 'after': after,
        'limit': limit,
      },
    );
    return MessagePage.fromJson(response.data!);
  }

  Future<String> uploadImage({
    required String fileName,
    required String mimeType,
    required Uint8List bytes,
  }) async {
    final response = await api.dio.post<Map<String, dynamic>>(
      '/media/uploads',
      data: {
        'file_name': fileName,
        'mime_type': mimeType,
        'size': bytes.length,
      },
    );
    final upload = response.data!;
    final headers = Map<String, dynamic>.from(upload['headers'] as Map);
    await Dio().put<void>(
      upload['upload_url'] as String,
      data: Stream.value(bytes),
      options: Options(headers: headers, contentType: mimeType),
    );
    return upload['upload_id'] as String;
  }

  Future<String> attachmentUrl(String attachmentId) async {
    final response = await api.dio.get<Map<String, dynamic>>(
      '/media/attachments/$attachmentId/download',
    );
    return response.data!['download_url'] as String;
  }
}
