## 2025-05-10 - Web Terminal Cancel Endpoint Lacks Authentication Check
**Vulnerability:** The `/cancel` endpoint allowed unauthenticated clients using any HTTP method (such as GET or POST) to kill the server's currently executing process.
**Learning:** Endpoints managing process lifecycles and background command cancellation must strictly validate HTTP methods and verify authenticated session credentials before taking destructive actions.
**Prevention:** Always perform explicit `r.Method != http.MethodPost` and `GetSessionUser(r)` checks on endpoints that modify execution state or terminate running processes.
