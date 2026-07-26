import 'package:web_socket_channel/web_socket_channel.dart';

WebSocketChannel connectWebSocket(Uri uri, String token) =>
    throw UnsupportedError(
      'WebSocket authentication is currently supported on native targets only.',
    );
