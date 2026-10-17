# Sentinel's Journal - Critical Security Learnings

## 2025-05-10 - Missing HTTP Server Timeouts (Slowloris DoS)
**Vulnerability:** Default Go `http.ListenAndServe` lacks read/write/header timeouts, leaving HTTP servers exposed to Slowloris resource exhaustion attacks.
**Learning:** `http.ListenAndServe` uses unbounded timeouts by default in Go's `net/http` package.
**Prevention:** Always construct an explicit `&http.Server{}` with `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, and `IdleTimeout` configured.
