import 'dart:async';

import 'package:equatable/equatable.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../core/api_client.dart';
import '../data/chat_repository.dart';
import '../models/models.dart';
import '../realtime/realtime_client.dart';

class ConversationsState extends Equatable {
  const ConversationsState({
    this.items = const [],
    this.loading = false,
    this.error,
  });
  final List<Conversation> items;
  final bool loading;
  final String? error;
  @override
  List<Object?> get props => [items, loading, error];
}

class ConversationsCubit extends Cubit<ConversationsState> {
  ConversationsCubit(this.repository, this.realtime)
    : super(const ConversationsState()) {
    _subscription = realtime.events
        .where((event) => event.type == 'conversation.updated')
        .listen((_) => load(silent: true));
  }
  final ChatRepository repository;
  final RealtimeClient realtime;
  late final StreamSubscription<RealtimeEvent> _subscription;

  Future<void> load({bool silent = false}) async {
    if (!silent) emit(ConversationsState(items: state.items, loading: true));
    try {
      emit(ConversationsState(items: await repository.conversations()));
    } catch (error) {
      emit(
        ConversationsState(items: state.items, error: readableApiError(error)),
      );
    }
  }

  Future<Conversation?> createDirect(User user) async {
    try {
      final conversation = await repository.createDirect(user.id);
      await load(silent: true);
      return conversation;
    } catch (error) {
      emit(
        ConversationsState(items: state.items, error: readableApiError(error)),
      );
      return null;
    }
  }

  @override
  Future<void> close() async {
    await _subscription.cancel();
    return super.close();
  }
}
