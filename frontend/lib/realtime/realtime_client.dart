import 'dart:async';
import 'dart:convert';

import '../core/app_config.dart';
import '../core/token_store.dart';
import 'realtime_connector.dart';

class RealtimeEvent {
  const RealtimeEvent(this.type, this.payload, {this.requestId = ''});
  final String type;
  final Map<String, dynamic> payload;
  final String requestId;
}

class RealtimeClient {
  RealtimeClient(this.tokens);
  final TokenStore tokens;
  final _events = StreamController<RealtimeEvent>.broadcast();
  Stream<RealtimeEvent> get events => _events.stream;
  dynamic _channel;
  Timer? _retry;
  int _attempt = 0;
  bool _closed = true;

  void connect() {
    if (tokens.accessToken == null) return;
    _closed = false;
    _retry?.cancel();
    try {
      _channel = connectWebSocket(AppConfig.websocketUri, tokens.accessToken!);
      _channel.stream.listen(
        _onData,
        onError: (_) => _reconnect(),
        onDone: _reconnect,
      );
      _attempt = 0;
      _events.add(const RealtimeEvent('connection.restored', {}));
    } catch (_) {
      _reconnect();
    }
  }

  void send(String type, String requestId, Map<String, dynamic> payload) {
    _channel?.sink.add(
      jsonEncode({'type': type, 'request_id': requestId, 'payload': payload}),
    );
  }

  void _onData(dynamic raw) {
    final json = jsonDecode(raw as String) as Map<String, dynamic>;
    _events.add(
      RealtimeEvent(
        json['type'] as String,
        Map<String, dynamic>.from(json['payload'] as Map? ?? {}),
        requestId: json['request_id'] as String? ?? '',
      ),
    );
  }

  void _reconnect() {
    if (_closed || _retry?.isActive == true) return;
    final seconds = (1 << _attempt.clamp(0, 5)).clamp(1, 30);
    _attempt++;
    _retry = Timer(Duration(seconds: seconds), connect);
  }

  Future<void> disconnect() async {
    _closed = true;
    _retry?.cancel();
    await _channel?.sink.close();
    _channel = null;
  }
}
