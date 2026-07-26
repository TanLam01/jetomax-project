# Realtime Chat Backend

Initial Go backend scaffold following the Clean Architecture boundaries in `plan requirement.md`.

## Requirements

- Go 1.26+
- PostgreSQL, Redis and MinIO (required as features are implemented)

## Run locally

```bash
cp .env.example .env
go mod download
go run ./cmd/api
```

With the existing local Docker containers `postgres17` and `redis`, PowerShell can discover their local credentials without writing secrets to disk:

```powershell
./scripts/run-local.ps1
```

The application automatically loads a local `.env`, while variables already exported by the process keep precedence. With `.env` configured, run directly:

```powershell
go run ./cmd/api
```

To run only the API in Docker and reuse the existing `postgres17` and `redis` containers exposed on the host:

```powershell
docker compose up --build -d
docker compose logs -f api
```

Docker Compose also starts MinIO for image storage and creates the configured bucket automatically:

- S3 API: `http://localhost:9000`
- MinIO Console: `http://localhost:9001`

MinIO data is persisted on the host at `backend/data/minio`. Recreating, stopping, or removing the MinIO container does not remove uploaded files. The directory is excluded from Git; back it up separately in production. Avoid deleting this directory when data must be retained.

`POST /api/v1/media/uploads` returns a pre-signed `PUT` URL. Upload the binary directly to that URL using exactly the returned headers before `expires_at`.

After the PUT succeeds, send the image through `/ws` using the returned `upload_id` (never trust or send an arbitrary object key):

```json
{
  "type": "message.send",
  "request_id": "request-001",
  "payload": {
    "conversation_id": "<conversation UUID>",
    "client_message_id": "<new UUID>",
    "message_type": "image",
    "text": "Optional caption",
    "upload_id": "<upload_id from POST /api/v1/media/uploads>"
  }
}
```

The backend verifies ownership and the object metadata in MinIO before creating the message attachment. An upload can only be attached to one message.

Conversation members resolve a private attachment to a 15-minute signed URL with:

```text
GET /api/v1/media/attachments/{attachmentId}/download
```

For reconnect synchronization, persist `sync_cursor` from the latest message response and call:

```text
GET /api/v1/conversations/{id}/messages?after={sync_cursor}&limit=100
```

The `after` response is oldest-first so it can be appended in order. Continue with `next_cursor` as the next `after` value while `has_more` is true, and persist the newest returned `sync_cursor`.

Supported WebSocket client events are `message.send`, `typing.start`, `typing.stop`, and `conversation.read`. Server events are `message.created`, `message.ack`, `presence.changed`, `typing.changed`, `conversation.updated`, and `error`.

The scaffold does not automatically load `.env`; export the variables in your shell or use your preferred dotenv runner. `DATABASE_URL` is required. `REDIS_URL` defaults to `redis://localhost:6379/0`.

- `GET http://localhost:8080/health/ready`
- `GET http://localhost:8080/api/v1/`
- `GET http://localhost:8080/swagger/index.html` — interactive Swagger UI

## Authentication API

- `POST /api/v1/auth/register` — `{ "email", "display_name", "password" }`
- `POST /api/v1/auth/login` — `{ "email", "password" }`
- `POST /api/v1/auth/refresh` — `{ "refresh_token" }`
- `POST /api/v1/auth/logout` — `{ "refresh_token" }`
- `GET /api/v1/users/me` — authenticated profile
- `GET /api/v1/users?query=` — authenticated user search (2-100 characters)
- `GET /api/v1/conversations` — member conversations with last message and unread count
- `POST /api/v1/conversations/direct` — create or return an existing direct conversation
- `POST /api/v1/conversations/groups` — create a group; creator becomes owner
- `POST /api/v1/conversations/{id}/members` — owner/admin adds group members
- `DELETE /api/v1/conversations/{id}/members/{userId}` — remove a member or leave a group
- `GET /api/v1/conversations/{id}/messages?cursor=&limit=` — cursor-paginated message history
- `POST /api/v1/media/uploads` — create a pre-signed JPEG/PNG/WebP upload URL
- `GET /ws` — authenticated WebSocket gateway (`message.send`, `message.ack`, `message.created`)

The users and conversations endpoints require `Authorization: Bearer <access_token>`.

Passwords are one-way hashed with bcrypt and are never returned by the API. Refresh tokens use rotation; only their SHA-256 hashes are stored. Access tokens are short-lived HS256 JWTs. Set a strong `JWT_ACCESS_SECRET` in every deployed environment.

## Error audit

Every HTTP `4xx`/`5xx` response receives an `X-Request-ID` and is recorded in both PostgreSQL table `errors` and the JSON Lines file configured by `ERROR_LOG_PATH` (default `logs/error.log`). Request/response bodies and credentials are deliberately excluded.

Client responses are intentionally generic by HTTP status (`Bad request`, `Unauthorized`, `Not found`, `Conflict`, `Internal server error`, etc.). The specific internal error code and message are available only in the server-side audit records and can be correlated using `X-Request-ID`.

Regenerate Swagger after changing an endpoint or DTO:

```bash
go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g cmd/api/main.go -o docs --parseInternal
```

## Architecture

- `cmd/api`: REST and WebSocket process entrypoint
- `cmd/mcp`: MCP process entrypoint
- `internal/domain`: framework-independent entities, errors and repository contracts
- `internal/usecase`: application orchestration grouped by feature
- `internal/delivery`: HTTP, WebSocket and MCP adapters
- `internal/delivery/http/dto`: typed HTTP request/response contracts and mappings
- `internal/infrastructure`: PostgreSQL/GORM, Redis, storage, AI and security adapters
- `internal/bootstrap`: dependency wiring
- `migrations`: versioned PostgreSQL migrations
- `tests`: integration and contract tests

Dependency direction is delivery/infrastructure -> usecase -> domain.
