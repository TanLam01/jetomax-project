param(
    [string]$PostgresContainer = "postgres17",
    [string]$RedisContainer = "redis",
    [string]$HttpPort = "8080"
)

$ErrorActionPreference = "Stop"

$postgres = docker inspect $PostgresContainer | ConvertFrom-Json
if (-not $postgres) {
    throw "PostgreSQL container '$PostgresContainer' was not found."
}

$redis = docker inspect $RedisContainer | ConvertFrom-Json
if (-not $redis) {
    throw "Redis container '$RedisContainer' was not found."
}
if (-not $postgres[0].State.Running -or -not $redis[0].State.Running) {
    throw "Both PostgreSQL and Redis containers must be running."
}

$postgresEnvironment = @{}
foreach ($item in $postgres[0].Config.Env) {
    $parts = $item -split "=", 2
    $postgresEnvironment[$parts[0]] = $parts[1]
}

$postgresUser = if ($postgresEnvironment["POSTGRES_USER"]) {
    $postgresEnvironment["POSTGRES_USER"]
} else {
    "postgres"
}
$postgresDatabase = if ($postgresEnvironment["POSTGRES_DB"]) {
    $postgresEnvironment["POSTGRES_DB"]
} else {
    $postgresUser
}
$postgresPassword = [uri]::EscapeDataString($postgresEnvironment["POSTGRES_PASSWORD"])

$env:DATABASE_URL = "postgres://${postgresUser}:${postgresPassword}@localhost:5432/${postgresDatabase}?sslmode=disable"
$env:REDIS_URL = "redis://localhost:6379/0"
$env:HTTP_PORT = $HttpPort

Write-Host "Starting API with PostgreSQL '$PostgresContainer' and Redis '$RedisContainer'..."
go run ./cmd/api
