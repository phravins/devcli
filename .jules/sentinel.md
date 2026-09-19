## 2025-04-04 - Unauthenticated Control Endpoints in Web Compiler
**Vulnerability:** The `/cancel` endpoint in `internal/web/server.go` allowed unauthenticated HTTP requests with any HTTP method to kill running server processes.
**Learning:** While execution endpoints (`/run`, `/terminal`, `/save`) checked `GetSessionUser(r)`, control endpoints like `/cancel` were omitted during authentication implementation.
**Prevention:** Always apply uniform authentication and HTTP method validation across all state-changing and process-control routes in `internal/web/server.go`.
