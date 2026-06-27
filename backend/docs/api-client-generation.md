# API Client / SDK Generation — Minisource Notifier

The Notifier API is documented via OpenAPI/Swagger. The generated spec is at `docs/swagger.json` / `docs/swagger.yaml`.

You can generate typed API clients for multiple languages from the OpenAPI spec.

---

## OpenAPI Spec Location

| Format | Path |
|--------|------|
| JSON | `docs/swagger.json` |
| YAML | `docs/swagger.yaml` |
| Swagger UI | `GET /swagger/index.html` (when server is running) |

## Regenerate Swagger

```bash
cd notifier/backend
swag init -g cmd/server/main.go --output docs
```

Or via Makefile:

```bash
make swagger
```

---

## TypeScript / Frontend Client

### Option 1: openapi-typescript (recommended for TypeScript frontends)

```bash
npm install -D openapi-typescript
npx openapi-typescript docs/swagger.yaml -o clients/typescript/notifier.ts
```

### Option 2: orval (React/Vue hooks)

```bash
npm install -D orval
npx orval --input docs/swagger.yaml --output clients/typescript/
```

### Option 3: openapi-generator (Java/Gradle)

```bash
npx @openapitools/openapi-generator-cli generate \
  -i docs/swagger.yaml \
  -g typescript-fetch \
  -o clients/typescript-fetch/
```

---

## Go Client (internal services)

### Option 1: oapi-codegen (echo/fiber compatible)

```bash
go install github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen@latest
oapi-codegen -package notifier -generate types,client docs/swagger.yaml > clients/go/notifier.gen.go
```

### Option 2: Manual HTTP client

For simple use cases, use `go-sdk/notifier` which provides a hand-crafted client:

```go
import "github.com/minisource/go-sdk/notifier"

client := notifier.NewClient("http://notifier:9002", "my-service-token")
resp, err := client.SendNotification(ctx, &notifier.SendRequest{
    UserID: userID,
    Type:   "email",
    Body:   "Hello",
})
```

---

## Postman Collection

An HTTP API collection is available at:

```txt
docs/http/notifier.http
```

Import this file into VS Code REST Client, or convert it to Postman format:

1. Open in VS Code with REST Client extension
2. Use individual requests
3. Set variables in `.env` or VS Code settings

---

## cURL Examples

### Health Check
```bash
curl http://localhost:9002/api/v1/health/
```

### Create Notification (service auth)
```bash
curl -X POST http://localhost:9002/api/v1/service/notifications \
  -H "Authorization: Bearer $SERVICE_TOKEN" \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: $(uuidgen)" \
  -d '{
    "userId": "123e4567-e89b-12d3-a456-426614174000",
    "type": "email",
    "body": "Hello from API"
  }'
```

### User Notifications
```bash
curl http://localhost:9002/api/v1/me/notifications \
  -H "Authorization: Bearer $USER_TOKEN"
```

### Admin Dashboard
```bash
curl http://localhost:9002/api/v1/admin/dashboard/overview \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

---

## Make Targets

```makefile
# Regenerate Swagger/OpenAPI spec
make swagger

# Generate TypeScript client
make client-typescript

# Validate OpenAPI spec
make validate-openapi
```
