# Jetomax Chat Flutter

Flutter client for the Golang REST/WebSocket backend. The first version targets Windows, Android and iOS with one adaptive codebase. Flutter Web compiles, but authenticated WebSocket is intentionally native-only because browsers cannot attach the required `Authorization` header.

## Run

Start the backend first, then from this directory:

```powershell
flutter pub get
flutter run -d windows
```

The default API is `http://localhost:8080/api/v1`. Override it with a compile-time value:

```powershell
flutter run -d windows --dart-define=API_BASE_URL=http://localhost:8080/api/v1
```

For the standard Android emulator, use the host alias:

```powershell
flutter run -d <android-device-id> --dart-define=API_BASE_URL=http://10.0.2.2:8080/api/v1
```

For a physical phone, use the development machine's LAN IP and allow the port through the firewall.

## Included

- Register, login, refresh rotation and logout
- Secure token storage
- User search and direct conversation creation
- Adaptive Messenger-style conversation/chat layout
- Cursor history and missed-message synchronization after reconnect
- Native authenticated WebSocket with exponential reconnect
- Text messages, typing, presence and read updates
- Image selection, direct MinIO upload and image messages
- Pending/sent message state

AI image, MCP and automation screens are outside this initial frontend scope.
