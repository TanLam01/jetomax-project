import 'dart:async';
import 'dart:typed_data';

import 'package:equatable/equatable.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:uuid/uuid.dart';

import '../core/api_client.dart';
import '../core/token_store.dart';
import '../data/chat_repository.dart';
import '../models/models.dart';
import '../realtime/realtime_client.dart';

class ChatState extends Equatable {
  const ChatState({
    this.messages = const [],
    this.loading = false,
    this.loadingOlder = false,
    this.hasMore = false,
    this.nextCursor,
    this.syncCursor,
    this.typingUsers = const {},
    this.peerOnline = false,
    this.error,
  });
  final List<Message> messages;
  final bool loading;
  final bool loadingOlder;
  final bool hasMore;
  final String? nextCursor;
  final String? syncCursor;
  final Set<String> typingUsers;
  final bool peerOnline;
  final String? error;
  @override
  List<Object?> get props => [
    messages,
    loading,
    loadingOlder,
    hasMore,
    nextCursor,
    syncCursor,
    typingUsers,
    peerOnline,
    error,
  ];
}

class ChatCubit extends Cubit<ChatState> {
  ChatCubit({
    required this.conversationId,
    required this.repository,
    required this.realtime,
    required this.tokens,
    this.peerUserId = '',
  }) : super(const ChatState()) {
    _subscription = realtime.events.listen(_onRealtime);
  }
  final String conversationId;
  final ChatRepository repository;
  final RealtimeClient realtime;
  final TokenStore tokens;
  final String peerUserId;
  late final StreamSubscription<RealtimeEvent> _subscription;
  final _uuid = const Uuid();
  Timer? _typingTimer;

  Future<void> load() async {
    emit(ChatState(messages: state.messages, loading: true));
    try {
      final page = await repository.messages(conversationId);
      emit(
        ChatState(
          messages: page.messages.reversed.toList(),
          hasMore: page.hasMore,
          nextCursor: page.nextCursor,
          syncCursor: page.syncCursor,
        ),
      );
      _markLatestRead();
    } catch (error) {
      emit(ChatState(messages: state.messages, error: readableApiError(error)));
    }
  }

  Future<void> loadOlder() async {
    if (!state.hasMore || state.loadingOlder || state.nextCursor == null) {
      return;
    }
    emit(_copy(loadingOlder: true));
    try {
      final page = await repository.messages(
        conversationId,
        cursor: state.nextCursor,
      );
      emit(
        _copy(
          messages: [...page.messages.reversed, ...state.messages],
          loadingOlder: false,
          hasMore: page.hasMore,
          nextCursor: page.nextCursor,
        ),
      );
    } catch (error) {
      emit(_copy(loadingOlder: false, error: readableApiError(error)));
    }
  }

  void sendText(String text) {
    final value = text.trim();
    if (value.isEmpty) return;
    _send(type: 'text', text: value);
  }

  Future<void> sendImage(
    String fileName,
    String mimeType,
    Uint8List bytes,
  ) async {
    try {
      final uploadId = await repository.uploadImage(
        fileName: fileName,
        mimeType: mimeType,
        bytes: bytes,
      );
      _send(type: 'image', text: '', uploadId: uploadId);
    } catch (error) {
      emit(_copy(error: readableApiError(error)));
    }
  }

  void _send({
    required String type,
    required String text,
    String? uploadId,
    String? retryClientId,
  }) {
    final clientId = retryClientId ?? _uuid.v4();
    final optimistic = Message(
      id: clientId,
      conversationId: conversationId,
      senderId: tokens.user!.id,
      type: type,
      text: text,
      clientMessageId: clientId,
      createdAt: DateTime.now(),
      status: MessageStatus.pending,
    );
    emit(
      _copy(
        messages: [
          ...state.messages.where((m) => m.clientMessageId != clientId),
          optimistic,
        ],
      ),
    );
    realtime.send('message.send', _uuid.v4(), {
      'conversation_id': conversationId,
      'client_message_id': clientId,
      'message_type': type,
      'text': text,
      if (uploadId != null) 'upload_id': uploadId,
    });
  }

  void typing() {
    realtime.send('typing.start', _uuid.v4(), {
      'conversation_id': conversationId,
    });
    _typingTimer?.cancel();
    _typingTimer = Timer(
      const Duration(seconds: 2),
      () => realtime.send('typing.stop', _uuid.v4(), {
        'conversation_id': conversationId,
      }),
    );
  }

  Future<void> _sync() async {
    if (state.syncCursor == null) return;
    try {
      var cursor = state.syncCursor;
      while (cursor != null) {
        final page = await repository.messages(
          conversationId,
          after: cursor,
          limit: 100,
        );
        final merged = _merge(state.messages, page.messages);
        emit(
          _copy(
            messages: merged,
            syncCursor: page.syncCursor ?? state.syncCursor,
          ),
        );
        cursor = page.hasMore ? page.nextCursor : null;
      }
      _markLatestRead();
    } catch (_) {}
  }

  void _onRealtime(RealtimeEvent event) {
    if (event.type == 'connection.restored') {
      _sync();
      return;
    }
    if (event.type == 'message.created' &&
        event.payload['conversation_id'] == conversationId) {
      final message = Message.fromJson(event.payload);
      emit(_copy(messages: _merge(state.messages, [message])));
      _markLatestRead();
    } else if (event.type == 'message.ack') {
      final clientId = event.payload['client_message_id'] as String?;
      emit(
        _copy(
          messages: state.messages
              .map(
                (message) => message.clientMessageId == clientId
                    ? message.copyWith(
                        id: event.payload['message_id'] as String?,
                        status: MessageStatus.sent,
                      )
                    : message,
              )
              .toList(),
        ),
      );
    } else if (event.type == 'typing.changed' &&
        event.payload['conversation_id'] == conversationId) {
      final users = {...state.typingUsers};
      final userId = event.payload['user_id'] as String;
      event.payload['is_typing'] == true
          ? users.add(userId)
          : users.remove(userId);
      emit(_copy(typingUsers: users));
    } else if (event.type == 'presence.changed' &&
        event.payload['user_id'] == peerUserId) {
      emit(_copy(peerOnline: event.payload['online'] == true));
    }
  }

  List<Message> _merge(List<Message> current, List<Message> incoming) {
    final byId = {for (final message in current) message.id: message};
    for (final message in incoming) {
      byId.removeWhere(
        (_, existing) =>
            existing.clientMessageId.isNotEmpty &&
            existing.clientMessageId == message.clientMessageId,
      );
      byId[message.id] = message;
    }
    final result = byId.values.toList()
      ..sort((a, b) => a.createdAt.compareTo(b.createdAt));
    return result;
  }

  void _markLatestRead() {
    if (state.messages.isEmpty) return;
    realtime.send('conversation.read', _uuid.v4(), {
      'conversation_id': conversationId,
      'message_id': state.messages.last.id,
    });
  }

  ChatState _copy({
    List<Message>? messages,
    bool? loading,
    bool? loadingOlder,
    bool? hasMore,
    String? nextCursor,
    String? syncCursor,
    Set<String>? typingUsers,
    bool? peerOnline,
    String? error,
  }) => ChatState(
    messages: messages ?? state.messages,
    loading: loading ?? false,
    loadingOlder: loadingOlder ?? state.loadingOlder,
    hasMore: hasMore ?? state.hasMore,
    nextCursor: nextCursor ?? state.nextCursor,
    syncCursor: syncCursor ?? state.syncCursor,
    typingUsers: typingUsers ?? state.typingUsers,
    peerOnline: peerOnline ?? state.peerOnline,
    error: error,
  );

  @override
  Future<void> close() async {
    _typingTimer?.cancel();
    await _subscription.cancel();
    return super.close();
  }
}
