import 'package:equatable/equatable.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../core/api_client.dart';
import '../core/token_store.dart';
import '../data/chat_repository.dart';
import '../models/models.dart';
import '../realtime/realtime_client.dart';

sealed class AuthState extends Equatable {
  const AuthState();
  @override
  List<Object?> get props => [];
}

class AuthLoading extends AuthState {
  const AuthLoading();
}

class AuthSignedOut extends AuthState {
  const AuthSignedOut();
}

class AuthSignedIn extends AuthState {
  const AuthSignedIn(this.user);
  final User user;
  @override
  List<Object?> get props => [user.id];
}

class AuthFailure extends AuthState {
  const AuthFailure(this.message);
  final String message;
  @override
  List<Object?> get props => [message];
}

class AuthCubit extends Cubit<AuthState> {
  AuthCubit(this.repository, this.tokens, this.realtime)
    : super(const AuthLoading()) {
    repository.api.onSessionExpired = () {
      realtime.disconnect();
      emit(const AuthSignedOut());
    };
  }
  final ChatRepository repository;
  final TokenStore tokens;
  final RealtimeClient realtime;

  Future<void> initialize() async {
    if (tokens.accessToken != null && tokens.user != null) {
      emit(AuthSignedIn(tokens.user!));
      realtime.connect();
    } else {
      emit(const AuthSignedOut());
    }
  }

  Future<void> login(String email, String password) =>
      _authenticate(() => repository.login(email.trim(), password));
  Future<void> register(String name, String email, String password) =>
      _authenticate(
        () => repository.register(name.trim(), email.trim(), password),
      );

  Future<void> _authenticate(Future<Session> Function() action) async {
    emit(const AuthLoading());
    try {
      final session = await action();
      await tokens.save(session);
      emit(AuthSignedIn(session.user));
      realtime.connect();
    } catch (error) {
      emit(AuthFailure(readableApiError(error)));
      emit(const AuthSignedOut());
    }
  }

  Future<void> logout() async {
    try {
      await repository.logout();
    } catch (_) {}
    await realtime.disconnect();
    await tokens.clear();
    emit(const AuthSignedOut());
  }
}
