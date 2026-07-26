class User {
  const User({
    required this.id,
    required this.email,
    required this.displayName,
    this.avatarKey = '',
  });
  final String id;
  final String email;
  final String displayName;
  final String avatarKey;

  factory User.fromJson(Map<String, dynamic> json) => User(
    id: json['id'] as String,
    email: json['email'] as String? ?? '',
    displayName: json['display_name'] as String? ?? '',
    avatarKey: json['avatar_key'] as String? ?? '',
  );
  Map<String, dynamic> toJson() => {
    'id': id,
    'email': email,
    'display_name': displayName,
    'avatar_key': avatarKey,
  };
}

class Session {
  const Session({
    required this.user,
    required this.accessToken,
    required this.refreshToken,
  });
  final User user;
  final String accessToken;
  final String refreshToken;
  factory Session.fromJson(Map<String, dynamic> json) => Session(
    user: User.fromJson(json['user'] as Map<String, dynamic>),
    accessToken: json['access_token'] as String,
    refreshToken: json['refresh_token'] as String,
  );
}

class Conversation {
  const Conversation({
    required this.id,
    required this.type,
    required this.name,
    required this.role,
    required this.unreadCount,
    this.peerUserId = '',
    this.lastMessage,
  });
  final String id;
  final String type;
  final String name;
  final String role;
  final int unreadCount;
  final String peerUserId;
  final Message? lastMessage;
  factory Conversation.fromJson(Map<String, dynamic> json) => Conversation(
    id: json['id'] as String,
    type: json['type'] as String,
    name: (json['name'] as String?)?.trim().isNotEmpty == true
        ? json['name'] as String
        : (json['type'] == 'group' ? 'Nhóm chat' : 'Tin nhắn'),
    role: json['role'] as String? ?? 'member',
    unreadCount: (json['unread_count'] as num?)?.toInt() ?? 0,
    peerUserId: json['peer_user_id'] as String? ?? '',
    lastMessage: json['last_message'] == null
        ? null
        : Message.fromJson(json['last_message'] as Map<String, dynamic>),
  );
}

class Attachment {
  const Attachment({
    required this.id,
    required this.objectKey,
    required this.mimeType,
    required this.size,
  });
  final String id;
  final String objectKey;
  final String mimeType;
  final int size;
  factory Attachment.fromJson(Map<String, dynamic> json) => Attachment(
    id: json['id'] as String,
    objectKey: json['object_key'] as String? ?? '',
    mimeType: json['mime_type'] as String? ?? '',
    size: (json['size'] as num?)?.toInt() ?? 0,
  );
}

class Message {
  const Message({
    required this.id,
    required this.conversationId,
    required this.senderId,
    required this.type,
    required this.clientMessageId,
    required this.createdAt,
    this.text = '',
    this.attachment,
    this.status = MessageStatus.sent,
  });
  final String id;
  final String conversationId;
  final String senderId;
  final String type;
  final String text;
  final String clientMessageId;
  final DateTime createdAt;
  final Attachment? attachment;
  final MessageStatus status;

  Message copyWith({
    String? id,
    Attachment? attachment,
    MessageStatus? status,
  }) => Message(
    id: id ?? this.id,
    conversationId: conversationId,
    senderId: senderId,
    type: type,
    text: text,
    clientMessageId: clientMessageId,
    createdAt: createdAt,
    attachment: attachment ?? this.attachment,
    status: status ?? this.status,
  );

  factory Message.fromJson(Map<String, dynamic> json) => Message(
    id: json['id'] as String,
    conversationId: json['conversation_id'] as String? ?? '',
    senderId: json['sender_id'] as String? ?? '',
    type: json['type'] as String? ?? 'text',
    text: json['text'] as String? ?? '',
    clientMessageId: json['client_message_id'] as String? ?? '',
    createdAt:
        DateTime.tryParse(json['created_at'] as String? ?? '') ??
        DateTime.now(),
    attachment: json['attachment'] == null
        ? null
        : Attachment.fromJson(json['attachment'] as Map<String, dynamic>),
  );
}

enum MessageStatus { pending, sent, failed }

class MessagePage {
  const MessagePage({
    required this.messages,
    this.nextCursor,
    this.syncCursor,
    required this.hasMore,
  });
  final List<Message> messages;
  final String? nextCursor;
  final String? syncCursor;
  final bool hasMore;
  factory MessagePage.fromJson(Map<String, dynamic> json) => MessagePage(
    messages: (json['data'] as List? ?? [])
        .map((item) => Message.fromJson(item as Map<String, dynamic>))
        .toList(),
    nextCursor: json['next_cursor'] as String?,
    syncCursor: json['sync_cursor'] as String?,
    hasMore: json['has_more'] as bool? ?? false,
  );
}
