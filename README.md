# Jetomax Real-time Chat

Jetomax is a real-time messaging application built as an interview assignment. It currently provides a Go REST/WebSocket backend and a Flutter client with a Messenger-inspired adaptive interface for Windows, Android, and iOS.

The repository is intentionally implemented as a modular monolith: it keeps local development and deployment simple while preserving clear boundaries between domain rules, application use cases, transport adapters, and infrastructure.

## Technology stack

### Backend

- **Language:** Go 1.26
- **HTTP:** Gin
- **ORM/database:** GORM and PostgreSQL 17
- **Real-time:** `coder/websocket` and Redis Pub/Sub
- **Object storage:** MinIO through the AWS S3 SDK
- **Authentication:** short-lived JWT access tokens and rotating refresh tokens
- **API documentation:** Swagger/OpenAPI
- **Containers:** Docker and Docker Compose

### Frontend

- **Language:** Dart 3.9
- **Framework:** Flutter 3.35
- **State management:** BLoC/Cubit
- **Networking:** Dio and `web_socket_channel`
- **Navigation:** `go_router`
- **Dependency injection:** GetIt
- **Secure credentials:** `flutter_secure_storage`
- **Images:** `file_picker` and `cached_network_image`

## Implemented features

### Backend

- Register, login, access-token refresh rotation, and logout
- One-way bcrypt password hashing; passwords are never returned
- User profile lookup and user search
- Direct conversation creation with duplicate prevention
- Group creation, adding members, removing members, and leaving a group
- Owner/admin/member authorization rules for existing group operations
- Conversation list ordered by recent activity, including last message and unread count
- Member-protected cursor-paginated message history
- Missed-message synchronization with an opaque `after` cursor
- Real-time text and image messages over authenticated WebSocket
- Idempotent sends using `client_message_id`
- Message acknowledgement, typing, presence, conversation update, and read events
- Redis Pub/Sub fan-out for multiple backend instances
- Multi-device presence tracking with heartbeat refresh
- MinIO pre-signed image upload and member-authorized pre-signed download
- Generic public API errors with detailed audit records in PostgreSQL and a JSON Lines log file
- Swagger UI
- Graceful HTTP, WebSocket, Redis, and database shutdown

### Flutter

- Register, login, automatic token refresh, and logout
- Secure token storage
- User search and direct chat creation
- Adaptive layout: one screen on mobile and conversation/chat split view on desktop
- Real-time text messages with pending/sent state
- Cursor history and missed-message synchronization after reconnect
- Exponential WebSocket reconnect
- Typing, presence, and read updates
- Image selection, direct MinIO upload, image messaging, and private image display
- Android and in-app back navigation

## Not implemented yet

- Flutter screens for creating groups and managing group members
- Backend endpoint for listing group members
- Promoting/demoting admins and transferring group ownership
- Complete profile editing and avatar management
- AI image-generation APIs and UI
- Functional MCP tools/server transport for ChatGPT
- n8n workflow, 24-hour summaries page, and Google Sheets integration
- Durable transactional outbox; the current Redis retry strategy is simpler and does not provide full delivery guarantees during prolonged outages
- Persistent Flutter offline cache/outbox with Drift or Isar
- Full WebSocket support for Flutter Web; authenticated WebSocket currently targets native platforms
- Complete integration, E2E, concurrent fan-out, and load-test suites
- Production deployment, observability, CI/CD, and backup automation

## Architecture and technical decisions

```text
Flutter
  ├── REST/Dio ────────────────┐
  └── WebSocket ───────────────┤
                               ▼
Gin HTTP / WebSocket delivery adapters
                               │
                               ▼
Application use cases (auth, users, conversations, messages, media)
                               │
                               ▼
Domain entities and repository contracts
                               ▲
                               │
GORM/PostgreSQL ─ Redis ─ MinIO infrastructure adapters
```

The main decisions are:

- **Clean Architecture boundaries:** delivery and infrastructure depend on use-case/domain contracts, while the domain does not import Gin, GORM, Redis, or AWS packages. This keeps business rules testable and transport-independent.
- **PostgreSQL as the source of truth:** users, membership, roles, messages, attachments, tokens, and error audits are durable. Redis is used only for ephemeral presence and cross-instance event fan-out.
- **REST plus WebSocket:** REST handles authentication, searchable resources, history, cursor synchronization, and media URLs. WebSocket handles low-latency events. This avoids forcing request/response workflows through a persistent socket.
- **Idempotent message sending:** the client creates a stable `client_message_id`; PostgreSQL enforces uniqueness per sender so reconnect/retry does not create duplicate messages.
- **Cursor pagination:** message ordering uses timestamp plus message UUID instead of offsets, avoiding skipped/duplicated pages when new messages arrive.
- **Private object storage:** clients upload directly using short-lived pre-signed PUT URLs. Download URLs are issued only after checking conversation membership, so the API does not proxy large image bodies.
- **Token rotation:** access tokens are short-lived. Refresh tokens rotate and only SHA-256 token hashes are stored, reducing the impact of a database leak.
- **Generic client errors and detailed server audits:** clients receive stable status-based messages without internal details; diagnostics are correlated by request ID in the `errors` table and log file.
- **Adaptive Flutter UI:** the same codebase uses a single-pane navigation flow on mobile and a two-pane conversation/chat layout on desktop.

## Repository layout

```text
.
├── backend/
│   ├── cmd/api/                 # REST and WebSocket entrypoint
│   ├── cmd/mcp/                 # Reserved MCP entrypoint
│   ├── internal/domain/         # Entities, errors, repository contracts
│   ├── internal/usecase/        # Application/business workflows
│   ├── internal/delivery/       # HTTP, WebSocket, MCP adapters
│   ├── internal/infrastructure/ # PostgreSQL, Redis, MinIO, security, logs
│   ├── migrations/              # Versioned PostgreSQL SQL migrations
│   ├── docs/                    # Generated Swagger files
│   └── docker-compose.yml       # API and persistent local MinIO
├── frontend/
│   ├── lib/core/                # Configuration, API client, token storage, DI
│   ├── lib/data/                # REST repository
│   ├── lib/models/              # API/application models
│   ├── lib/realtime/            # Native WebSocket connection
│   ├── lib/bloc/                # Auth, conversation, and chat state
│   └── lib/ui/                  # Adaptive screens and widgets
├── plan requirement.md
└── assignment_overview.docx
```

## Prerequisites

- Git
- Docker Desktop
- Go 1.26 or the Go version declared in `backend/go.mod`
- Flutter 3.35+ with Dart 3.9+
- Visual Studio with **Desktop development with C++** for Windows builds
- Android Studio/SDK and an emulator for Android development
- Xcode on macOS for iOS builds

Windows Flutter builds using plugins require Developer Mode:

```powershell
start ms-settings:developers
```

Enable **Developer Mode**, then reopen the terminal.

## Quick start

The commands below use PowerShell and expect PostgreSQL and Redis to be exposed on the host at ports `5432` and `6379`.

### 1. Start PostgreSQL and Redis

Skip this step if the existing `postgres17` and `redis` containers are already running.

```powershell
docker run -d `
  --name postgres17 `
  -e POSTGRES_USER=chat `
  -e POSTGRES_PASSWORD=chat `
  -e POSTGRES_DB=chat `
  -p 5432:5432 `
  -v jetomax-postgres:/var/lib/postgresql/data `
  postgres:17-alpine

docker run -d `
  --name redis `
  -p 6379:6379 `
  -v jetomax-redis:/data `
  redis:7-alpine redis-server --appendonly yes
```

Use stronger credentials outside local development.

### 2. Configure the backend

```powershell
cd backend
Copy-Item .env.example .env
```

Update `backend/.env` so it matches the PostgreSQL and MinIO credentials:

```env
APP_ENV=development
HTTP_HOST=0.0.0.0
HTTP_PORT=8080
DATABASE_URL=postgres://chat:chat@localhost:5432/chat?sslmode=disable
REDIS_URL=redis://localhost:6379/0
JWT_ACCESS_SECRET=replace-with-a-long-random-secret
ACCESS_TOKEN_TTL=15m
REFRESH_TOKEN_TTL=720h
ERROR_LOG_PATH=logs/error.log
S3_ENDPOINT=http://localhost:9000
S3_REGION=us-east-1
S3_BUCKET=chat-media
S3_ACCESS_KEY=minio
S3_SECRET_KEY=replace-with-a-minio-password
S3_USE_PATH_STYLE=true
```

The backend automatically loads `backend/.env`. Environment variables already set in the process take precedence.

### 3. Start MinIO and create the bucket

From `backend/`:

```powershell
docker compose up -d minio minio-init
```

- MinIO API: <http://localhost:9000>
- MinIO console: <http://localhost:9001>
- Persistent files: `backend/data/minio/`

### 4. Apply PostgreSQL migrations

The project currently uses versioned SQL files without an embedded migration runner. Apply every `*.up.sql` file in order for a fresh database:

```powershell
$pgUser = (docker exec postgres17 printenv POSTGRES_USER).Trim()
$pgDatabase = docker exec postgres17 printenv POSTGRES_DB
if ([string]::IsNullOrWhiteSpace($pgDatabase)) { $pgDatabase = $pgUser }
$pgDatabase = $pgDatabase.Trim()

Get-ChildItem migrations\*.up.sql |
  Sort-Object Name |
  ForEach-Object {
    Write-Host "Applying $($_.Name)"
    Get-Content -Raw $_.FullName |
      docker exec -i postgres17 psql -v ON_ERROR_STOP=1 -U $pgUser -d $pgDatabase
  }
```

Only run the full sequence against a fresh database. Existing databases should receive only migrations that have not already been applied.

### 5. Run the backend

From `backend/`:

```powershell
go mod download
go run ./cmd/api
```

If the existing containers use credentials you do not want to copy into `.env`, the helper can discover them for the current process:

```powershell
.\scripts\run-local.ps1
```

Verify:

- Readiness: <http://localhost:8080/health/ready>
- Swagger: <http://localhost:8080/swagger/index.html>
- REST base URL: `http://localhost:8080/api/v1`
- WebSocket: `ws://localhost:8080/ws`

### 6. Run Flutter on Windows

Open another terminal:

```powershell
cd frontend
flutter pub get
flutter run -d windows `
  --dart-define=API_BASE_URL=http://localhost:8080/api/v1
```

### 7. Run Flutter on an Android emulator

Android emulator uses `10.0.2.2` to access the Windows host:

```powershell
cd frontend
flutter devices
flutter run -d emulator-5554 `
  --dart-define=API_BASE_URL=http://10.0.2.2:8080/api/v1
```

Stop and rerun the app after changing `API_BASE_URL`; hot reload cannot change a compile-time Dart definition.

For a physical phone, use the computer's LAN address, for example:

```powershell
flutter run -d <device-id> `
  --dart-define=API_BASE_URL=http://192.168.1.10:8080/api/v1
```

Allow inbound TCP port `8080` through the development machine firewall when using a physical device.

## Optional database web UI

PostgreSQL port `5432` is not an HTTP website. Run Adminer to inspect it in a browser:

```powershell
docker run -d `
  --name jetomax-adminer `
  -p 8081:8080 `
  --add-host=host.docker.internal:host-gateway `
  adminer
```

Open <http://localhost:8081> and use:

```text
System:   PostgreSQL
Server:   host.docker.internal
Username: POSTGRES_USER
Password: POSTGRES_PASSWORD
Database: POSTGRES_DB
```

## Basic usage

1. Run the backend and Flutter client.
2. Register two different accounts.
3. Sign in as the first user and search for the second user.
4. Create a direct conversation.
5. Keep the second account connected on another emulator/device/window.
6. Send text or select an image from the composer.
7. Observe real-time delivery, acknowledgement, typing, presence, and unread/read updates.

Group messaging already uses the same message history and WebSocket flow after a group exists. The Flutter create/manage-group screens are still pending.

## API and WebSocket

Swagger documents the REST contract at <http://localhost:8080/swagger/index.html>.

The WebSocket handshake requires:

```text
Authorization: Bearer <access_token>
```

Client event types:

```text
message.send
typing.start
typing.stop
conversation.read
```

Server event types:

```text
message.created
message.ack
presence.changed
typing.changed
conversation.updated
error
```

Every server event contains `type`, `event_id`, `timestamp`, `request_id`, and `payload`.

## Test, analyze, and build

Backend:

```powershell
cd backend
go test ./...
go build ./cmd/api
```

Regenerate Swagger after changing handlers or DTOs:

```powershell
go run github.com/swaggo/swag/cmd/swag@v1.16.6 init `
  -g cmd/api/main.go -o docs
```

Flutter:

```powershell
cd frontend
dart format --output=none --set-exit-if-changed lib test
flutter analyze
flutter test
flutter build windows
```

## Troubleshooting

### Android reports that it cannot connect to the server

Do not use `localhost` from the emulator. Rerun Flutter with:

```powershell
--dart-define=API_BASE_URL=http://10.0.2.2:8080/api/v1
```

Confirm the backend is available from Windows at <http://localhost:8080/health/ready>.

### Port 8080 is already in use

Find the owning process:

```powershell
Get-NetTCPConnection -LocalPort 8080 -State Listen |
  Select-Object LocalAddress, LocalPort, OwningProcess
```

Stop the old API process or change `HTTP_PORT` and the Flutter `API_BASE_URL` together.

### Windows build says plugins require symlink support

Enable Windows Developer Mode:

```powershell
start ms-settings:developers
```

Then restart the terminal and run `flutter build windows` again.

### Android emulator focuses an input but does not show its software keyboard

Enable **Settings → System → Keyboard → Physical keyboard → Show virtual keyboard** inside the emulator.

### Image upload fails

- Confirm MinIO is healthy and the configured bucket exists.
- Ensure the file is JPEG, PNG, or WebP and no larger than 10 MB.
- Upload using exactly the headers returned by `POST /api/v1/media/uploads`.
- Finish the PUT before sending the returned `upload_id` through WebSocket.

### Detailed API errors are not returned to the client

This is intentional. Use `X-Request-ID` to correlate the error with:

- PostgreSQL table `errors`
- `backend/logs/error.log`

## Security notes

- Never commit `backend/.env`, access tokens, refresh tokens, database passwords, or MinIO credentials.
- Replace all example secrets before deployment.
- Do not expose Adminer or the MinIO console publicly.
- Use HTTPS/WSS and secure cookies or an equivalent protected token transport in production.
- Restrict production CORS/origin rules and rotate compromised credentials.

## Additional documentation

- [Backend details](backend/README.md)
- [Flutter details](frontend/README.md)
- [Implementation plan](plan%20requirement.md)
