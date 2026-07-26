import 'dart:convert';

import 'package:flutter_secure_storage/flutter_secure_storage.dart';

import '../models/models.dart';

class TokenStore {
  TokenStore(this._storage);

  static const _accessKey = 'access_token';
  static const _refreshKey = 'refresh_token';
  static const _userKey = 'current_user';

  final FlutterSecureStorage _storage;

  String? accessToken;
  String? refreshToken;
  User? user;

  Future<void> load() async {
    accessToken = await _storage.read(key: _accessKey);
    refreshToken = await _storage.read(key: _refreshKey);
    final rawUser = await _storage.read(key: _userKey);
    if (rawUser != null) user = User.fromJson(jsonDecode(rawUser));
  }

  Future<void> save(Session session) async {
    accessToken = session.accessToken;
    refreshToken = session.refreshToken;
    user = session.user;
    await Future.wait([
      _storage.write(key: _accessKey, value: accessToken),
      _storage.write(key: _refreshKey, value: refreshToken),
      _storage.write(key: _userKey, value: jsonEncode(user!.toJson())),
    ]);
  }

  Future<void> clear() async {
    accessToken = null;
    refreshToken = null;
    user = null;
    await Future.wait([
      _storage.delete(key: _accessKey),
      _storage.delete(key: _refreshKey),
      _storage.delete(key: _userKey),
    ]);
  }
}
