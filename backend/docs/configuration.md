# Configuration — Minisource Notifier

All configuration is via environment variables. See `.env.example` for a template.

---

## Server

| Variable | Required | Default | Description | Production |
|----------|----------|---------|-------------|------------|
| `SERVER_INTERNAL_PORT` | Yes | `9002` | Internal HTTP server port | Match Docker EXPOSE |
| `SERVER_EXTERNAL_PORT` | Yes | `9002` | External port (from PORT env or proxy) | Set by platform |
| `SERVER_RUN_MODE` | Yes | `development` | `development`, `staging`, `production` | `production` |
| `SERVER_NAME` | No | `Notifier` | Service name (used in logs/headers) | Set per environment |

## Database (PostgreSQL)

| Variable | Required | Default | Description | Production |
|----------|----------|---------|-------------|------------|
| `POSTGRES_HOST` | Yes | `localhost` | Database hostname | Internal DB host |
| `POSTGRES_PORT` | Yes | `5432` | Database port | `5432` |
| `POSTGRES_USER` | Yes | `postgres` | Database user | Least-privilege user |
| `POSTGRES_PASSWORD` | Yes | `postgres` | Database password | Strong secret |
| `POSTGRES_DBNAME` | Yes | `notifier_db` | Database name | `notifier_prod` |
| `POSTGRES_SSLMODE` | No | `disable` | SSL mode: `disable`, `require`, `verify-full` | `require` or `verify-full` |
| `POSTGRES_MAX_IDLE_CONNS` | No | `10` | Max idle connections | 10-25 |
| `POSTGRES_MAX_OPEN_CONNS` | No | `100` | Max open connections | 25-100 |
| `POSTGRES_CONN_MAX_LIFETIME` | No | `60` | Connection max lifetime (seconds) | 300-600 |

## Auth

| Variable | Required | Default | Description | Production |
|----------|----------|---------|-------------|------------|
| `AUTH_ENABLED` | No | `true` | Enable JWT auth middleware | `true` |
| `AUTH_BASE_URL` | No | `http://localhost:9001` | Auth service URL | Internal auth URL |
| `AUTH_CLIENT_ID` | Conditional | `` | Service client ID for auth | Required if AUTH_ENABLED |
| `AUTH_CLIENT_SECRET` | Conditional | `` | Service client secret | Required if AUTH_ENABLED |
| `AUTH_JWT_SECRET` | Conditional | `` | JWT secret for token validation | Required if AUTH_ENABLED |

## Rate Limiting

| Variable | Required | Default | Description | Production |
|----------|----------|---------|-------------|------------|
| `RATE_LIMIT_ENABLED` | No | `true` | Enable rate limiter | `true` |
| `RATE_LIMIT_REQUESTS` | No | `100` | Requests per window | 100-1000 |
| `RATE_LIMIT_WINDOW_SECONDS` | No | `60` | Rate limit window in seconds | 60 |
| `RATE_LIMIT_PROVIDER_TEST_REQUESTS` | No | `10` | Provider test limit per window | 5-10 |
| `RATE_LIMIT_NOTIFICATION_CREATE_REQUESTS` | No | `30` | Notification create limit per window | 30-100 |

## CORS

| Variable | Required | Default | Description | Production |
|----------|----------|---------|-------------|------------|
| `CORS_ALLOW_ORIGINS` | Yes | `*` | Allowed CORS origins | Specific domain(s) |
| `CORS_ALLOW_METHODS` | No | `GET,POST,PUT,PATCH,DELETE,OPTIONS` | Allowed HTTP methods | Default |
| `CORS_ALLOW_HEADERS` | No | `Origin,Content-Type,Accept,Authorization,X-Request-Id,X-Tenant-Id` | Allowed headers | Default |
| `CORS_ALLOW_CREDENTIALS` | No | `false` | Allow credentials in CORS | Depends on frontend |

## Worker

| Variable | Required | Default | Description | Production |
|----------|----------|---------|-------------|------------|
| `WORKER_NUM_WORKERS` | No | `10` | Number of concurrent workers | 10-50 |
| `WORKER_QUEUE_SIZE` | No | `1000` | In-memory queue buffer size | 1000-5000 |
| `WORKER_RETRY_MAX_DELAY` | No | `300` | Max retry delay in seconds | 300 |
| `WORKER_RETRY_BASE_DELAY` | No | `5` | Base retry delay in seconds | 5-30 |
| `WORKER_POLL_ENABLED` | No | `true` | Enable DB polling for pending | `true` |
| `WORKER_POLL_INTERVAL` | No | `15` | DB poll interval in seconds | 15-60 |

## Logger

| Variable | Required | Default | Description | Production |
|----------|----------|---------|-------------|------------|
| `LOGGER_FILE_PATH` | No | `logs/notifier.log` | Log file path | stdout via Docker |
| `LOGGER_ENCODING` | No | `json` | Log encoding: `json` or `console` | `json` |
| `LOGGER_LEVEL` | No | `info` | Log level: `debug`, `info`, `warn`, `error` | `info` or `warn` |
| `LOGGER_TYPE` | No | `zap` | Logger implementation | `zap` |
| `LOGGER_CONSOLE_ONLY` | No | `false` | Console-only logging | `true` for Docker |

## gRPC

| Variable | Required | Default | Description | Production |
|----------|----------|---------|-------------|------------|
| `GRPC_PORT` | No | `9003` | gRPC server port | 9003 |
| `GRPC_ENABLED` | No | `true` | Enable gRPC server | `true` |

## Database Behavior

| Variable | Required | Default | Description | Production |
|----------|----------|---------|-------------|------------|
| `DB_RUN_MIGRATIONS` | No | `false` | Run SQL migrations on startup | `true` (with caution) |
| `DB_RUN_SEED_DATA` | No | `false` | Run seed data on startup | `false` |
| `DB_AUTO_MIGRATE` | No | `false` | GORM AutoMigrate (dev only) | `false` |

## Providers

| Variable | Required | Default | Description | Production |
|----------|----------|---------|-------------|------------|
| `KAVENEGAR_ENABLED` | No | `false` | Enable Kavenegar SMS | Set if using Kavenegar |
| `KAVENEGAR_API_KEY` | Conditional | `` | Kavenegar API key | Strong secret |
| `KAVENEGAR_TEMPLATE` | No | `verify` | Default Kavenegar template | As configured |

## Tracing (Jaeger)

| Variable | Required | Default | Description | Production |
|----------|----------|---------|-------------|------------|
| `TRACING_ENABLED` | No | `false` | Enable Jaeger tracing | `true` if Jaeger available |
| `JAEGER_URL` | Conditional | `http://localhost:14268/api/traces` | Jaeger collector URL | Internal Jaeger URL |
| `TRACING_SERVICE_NAME` | No | `notifier-service` | Service name in traces | Per environment |

## Digest

| Variable | Required | Default | Description | Production |
|----------|----------|---------|-------------|------------|
| `DIGEST_ENABLED` | No | `true` | Enable digest accumulation | `true` |
| `DIGEST_INTERVAL` | No | `60` | Digest processing interval (seconds) | 60-300 |
| `DIGEST_BATCH_SIZE` | No | `50` | Max notifications per digest | 50-200 |
| `DIGEST_MAX_BODY_LEN` | No | `200` | Max body length per digest item | 200-500 |

## Secret Management

For production:
- Never commit `.env` files with real secrets
- Use Docker secrets, Kubernetes secrets, or a vault (HashiCorp Vault, AWS Secrets Manager)
- Rotate passwords and API keys regularly
- Use strong, random values for `AUTH_JWT_SECRET` and `AUTH_CLIENT_SECRET`
