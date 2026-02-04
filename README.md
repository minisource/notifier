# Notifier Service

A high-performance, scalable notification microservice that supports multiple notification channels including SMS, Email, Push notifications, and In-App notifications with WebSocket support.

## Features

- **Multi-Channel Support**: SMS, Email, Push, and In-App notifications
- **Database-Driven Configuration**: Store notification providers and settings in PostgreSQL
- **Retry Mechanism**: Exponential backoff retry policy for failed notifications
- **Real-Time Notifications**: WebSocket support for instant in-app notifications
- **Notification Templates**: Reusable templates with variable substitution
- **User Preferences**: Per-user notification preferences management
- **Batch Processing**: Send notifications to multiple recipients efficiently
- **Worker Queue**: Asynchronous processing with configurable worker pool
- **Notification History**: Track and query notification history with detailed logs
- **High Scalability**: Built for high-traffic microservice environments

## Architecture

- **API Layer**: RESTful APIs with Fiber framework
- **Service Layer**: Business logic and notification orchestration
- **Repository Layer**: Database operations with GORM
- **Worker Pool**: Background processing with retry logic
- **WebSocket Hub**: Real-time notification broadcasting
- **Database**: PostgreSQL with optimized indexes

## Getting Started

### Prerequisites

- Go 1.23.4 or later
- PostgreSQL 16 or later
- Docker (optional, for containerized deployment)

### Installation

1. Clone the repository:
```sh
git clone https://github.com/minisource/notifier.git
cd notifier
```

2. Install dependencies:
```sh
go mod download
```

### Configuration

The service supports multiple environments: `development`, `docker`, and `production`.

Set the `APP_ENV` environment variable:
```sh
export APP_ENV=development  # or docker, production
```

Configuration files are located in `config/`:
- `config-development.yml` - Local development
- `config-docker.yml` - Docker environment
- `config-production.yml` - Production environment

#### Configuration Structure

```yaml
server:
  internalPort: 9002
  externalPort: 9002
  name: "Notifier"

postgres:
  host: "localhost"
  port: "5432"
  user: "postgres"
  password: "postgres"
  dbName: "notifier_db"
  sslMode: "disable"
  maxIdleConns: 10
  maxOpenConns: 100
  connMaxLifetime: 60

worker:
  numWorkers: 10      # Number of concurrent workers
  queueSize: 1000     # Queue buffer size
  retryMaxDelay: 300  # Max retry delay in seconds
  retryBaseDelay: 5   # Base retry delay in seconds

sms:
  notEnabled: false
  defaultProvider: "kavenegar"
  providers:
    - provider: "kavenegar"
      apiKey: "your-api-key"
      template: "your-template"

email:
  notEnabled: false
  defaultProvider: "smtp"
  providers:
    - provider: "smtp"
      host: "smtp.gmail.com"
      port: 587
      username: "your-email@gmail.com"
      password: "your-app-password"
      from: "noreply@yourapp.com"
```

### Running the Application

#### Local Development

```sh
go run cmd/server/main.go
```

#### Using Docker Compose

```sh
docker-compose up -d
```

This will start:
- PostgreSQL database on port 5432
- Notifier service on port 9002

#### Using Taskfile

```sh
task run
```

## API Endpoints

### Health Check
```
GET /api/v1/health
```

### Notifications

#### Create Notification
```
POST /api/v1/notifications
Content-Type: application/json

{
  "userId": "uuid",
  "type": "sms|email|push|in_app",
  "priority": "low|normal|high|urgent",
  "recipientPhone": "+1234567890",
  "recipientEmail": "user@example.com",
  "subject": "Subject",
  "body": "Message body",
  "templateId": "uuid (optional)",
  "metadata": {},
  "scheduledAt": "2026-02-01T10:00:00Z (optional)"
}
```

#### Create Batch Notifications
```
POST /api/v1/notifications/batch
Content-Type: application/json

{
  "notifications": [...]
}
```

#### Get User Notifications
```
GET /api/v1/notifications/user/{userId}?page=1&pageSize=20
```

#### Get Unread Notifications
```
GET /api/v1/notifications/user/{userId}/unread?page=1&pageSize=20
```

#### Mark as Read
```
PUT /api/v1/notifications/{notificationId}/read
```

### Notification Preferences

#### Get User Preferences
```
GET /api/v1/preferences/user/{userId}
```

#### Update Preference
```
PUT /api/v1/preferences/user/{userId}
Content-Type: application/json

{
  "type": "sms|email|push|in_app",
  "isEnabled": true,
  "allowInstant": true,
  "allowDigest": false,
  "digestFrequency": "daily",
  "categorySettings": {
    "marketing": false,
    "alerts": true
  }
}
```

### Notification Templates

#### Create Template
```
POST /api/v1/templates
Content-Type: application/json

{
  "name": "welcome-email",
  "type": "email",
  "subject": "Welcome {{userName}}",
  "body": "Hello {{userName}}, welcome to our service!",
  "description": "Welcome email template",
  "variables": ["userName"],
  "provider": "smtp"
}
```

#### Get All Templates
```
GET /api/v1/templates?page=1&pageSize=50
```

#### Get Template
```
GET /api/v1/templates/{templateId}
```

#### Update Template
```
PUT /api/v1/templates/{templateId}
```

#### Delete Template
```
DELETE /api/v1/templates/{templateId}
```

### WebSocket

#### Connect to WebSocket
```
ws://localhost:9002/ws?userId={uuid}
```

WebSocket messages format:
```json
{
  "type": "notification",
  "data": {
    "id": "uuid",
    "userId": "uuid",
    "type": "in_app",
    "subject": "New message",
    "body": "You have a new message",
    "createdAt": "2026-01-31T12:00:00Z"
  },
  "time": "2026-01-31T12:00:00Z"
}
```

## Database Configuration from Settings

The service can load SMS and Email provider configurations from the database `settings` table, allowing dynamic configuration without redeployment.

### Settings Table Structure

| Key | Value | Category |
|-----|-------|----------|
| `sms.providers` | JSON array of SMS providers | `sms` |
| `sms.default_provider` | Default SMS provider name | `sms` |
| `email.providers` | JSON array of Email providers | `email` |
| `email.default_provider` | Default Email provider name | `email` |
| `notification.max_retries` | Maximum retry attempts | `notification` |
| `worker.pool_size` | Worker pool size | `worker` |

The service will check the database first, then fall back to config files if settings are not found.

## Retry Policy

The service implements an exponential backoff retry strategy:

1. **Initial Failure**: Notification fails
2. **Retry Schedule**: 
   - Retry 1: 5 seconds (base delay)
   - Retry 2: 10 seconds (2^1 * base)
   - Retry 3: 20 seconds (2^2 * base)
   - Max delay: 300 seconds (5 minutes)
3. **Max Retries**: Configurable (default: 3)
4. **Final State**: Marked as `failed` after max retries

## Worker Pool

- Configurable number of concurrent workers
- Queue-based processing with buffer
- Priority-based notification processing
- Periodic retry processor (runs every 30 seconds)
- Graceful shutdown with pending job completion

## Monitoring

The service includes detailed logging using the common_go logger:

- **Request/Response logs**: All API calls
- **Database operations**: Queries and performance
- **Worker activity**: Job processing and retries
- **WebSocket connections**: Client connections and broadcasts
- **Error tracking**: Detailed error logs with context

## Production Considerations

1. **Database Connection Pool**: Configure based on expected load
2. **Worker Pool Size**: Tune based on throughput requirements
3. **Queue Size**: Ensure sufficient buffer for traffic spikes
4. **Retry Strategy**: Adjust delays based on provider SLAs
5. **WebSocket Connections**: Monitor connection count
6. **Database Indexes**: Already optimized for common queries
7. **Logging**: Use structured logging for better observability

## Environment Variables

- `APP_ENV`: Environment (development, docker, production)
- `PORT`: Override external port
- `POSTGRES_HOST`: Database host
- `POSTGRES_PORT`: Database port
- `POSTGRES_USER`: Database user
- `POSTGRES_PASSWORD`: Database password
- `POSTGRES_DBNAME`: Database name
- `POSTGRES_SSLMODE`: SSL mode (disable, require)

## License

MIT License

## Contributing

Contributions are welcome! Please submit pull requests or open issues for bugs and feature requests.

