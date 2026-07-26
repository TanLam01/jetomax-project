import 'package:flutter_test/flutter_test.dart';
import 'package:jetomax_chat/models/models.dart';

void main() {
  test('parses an image message attachment', () {
    final message = Message.fromJson({
      'id': 'message-1',
      'conversation_id': 'conversation-1',
      'sender_id': 'user-1',
      'type': 'image',
      'client_message_id': 'client-1',
      'created_at': '2026-01-01T00:00:00Z',
      'attachment': {
        'id': 'attachment-1',
        'object_key': 'users/u/images/a.jpg',
        'mime_type': 'image/jpeg',
        'size': 128,
      },
    });

    expect(message.type, 'image');
    expect(message.attachment?.mimeType, 'image/jpeg');
  });
}
