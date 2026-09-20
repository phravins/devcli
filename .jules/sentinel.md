## 2025-05-18 - Missing Session Auth and Method Enforcement on Process Control Endpoints
**Vulnerability:** The `/cancel` endpoint in `internal/web/server.go` allowed unauthenticated users to cancel running process executions via any HTTP method.
**Learning:** In HTTP-based CLI/IDE web modules, state-changing and process-control endpoints (`/run`, `/terminal`, `/save`, `/cancel`) must strictly enforce both HTTP `POST` method validation and `GetSessionUser(r)` session authentication.
**Prevention:** Always verify session authentication (`GetSessionUser(r)`) and method (`r.Method == http.MethodPost`) on all non-public routes in `internal/web/server.go`.
