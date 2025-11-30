# Projects to do
- Set up and Data Model
  - Models
    - User 
    - Sessions
    - Device
  - Use PostgreSQL DB connections
    - GORM
- Registration API endpoint (done)
- User Login API (done)
- Secure Session Creation (done)
  - Upon successful login: generate a cryptographically secure, unique Session ID (e.g., UUID). Store the Session ID, User ID, Expiration Time, and Device Info in Redis.
- Device Tracking Logic
  - When creating a session, extract and store basic device/client info (e.g., User-Agent string, IP address) and link it to the session record in Redis.
- Logout / revoke session (done)
  - Implement a /logout endpoint. Delete the Session ID key from Redis to immediately invalidate the session and token.
- Fetch user profile api (done)
  - Implement a /profile endpoint that requires a valid session token (middleware). Fetch and return basic user data from PostgreSQL.
---
- Session Validation Middleware (done)
  - Create a Go HTTP middleware that runs on all protected routes (like /profile, /logout). The middleware must check the session token against Redis for quick lookups.
- Token Bucket Rate Limiter (done)
  - Implement a server-side token bucket rate limiter specifically for the /login endpoint to prevent brute-force attacks. Store the rate limit state (token count, last refill time) in-memory or in Redis and protect access using Go's sync.Mutex or a channel pattern for concurrency.
- Go Concurrency for Logs
  - Use a Goroutine and a Channel pattern to asynchronously log login attempts (success and failure) to a file or database after the request has been handled. This ensures the logging doesn't block the main request handler.
- Connection pooling
  - Configure and use robust connection pooling for both PostgreSQL and Redis to handle a large number of concurrent connections efficiently.
- Graceful Shutdown
  - Implement logic to shut down the server gracefully using Go's context package. This includes waiting for active requests to finish and properly closing database/Redis connections.
- JWT sessions
---
### Potential Struct
```
account-service-go/
├── api/             # HTTP handlers and request/response structs
│   ├── handlers.go
│   └── middleware.go
├── internal/        # Core business logic (e.g., service interfaces)
│   ├── account/
│   │   └── service.go
│   └── auth/
│       └── service.go
├── pkg/             # Reusable components (e.g., rate limiter, security utils)
│   ├── ratelimit/
│   │   └── token_bucket.go
│   └── util/
│       └── security.go (bcrypt helper)
├── db/              # Database initialization and migration scripts
├── main.go          # Entry point, server setup, config loading
└── go.mod
```

api -> models ->