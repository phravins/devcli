# Sentinel's Journal - Critical Security Learnings

## 2025-05-10 - Missing HTTP Server Timeouts (Slowloris DoS)
**Vulnerability:** Default Go `http.ListenAndServe` lacks read/write/header timeouts, leaving HTTP servers exposed to Slowloris resource exhaustion attacks.
**Learning:** `http.ListenAndServe` uses unbounded timeouts by default in Go's `net/http` package.
**Prevention:** Always construct an explicit `&http.Server{}` with `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, and `IdleTimeout` configured.

## 2026-10-17 - Missing Outbound HTTP Client Timeouts and Unbounded Response Reads
**Vulnerability:** Outbound HTTP requests created with bare `&http.Client{}` lack timeouts, causing goroutines to hang indefinitely if remote servers stall. Additionally, unbounded `io.ReadAll(resp.Body)` allows remote servers or MITM proxies to trigger memory exhaustion DoS.
**Learning:** Go's `http.DefaultClient` and `&http.Client{}` have zero timeout by default, and `io.ReadAll` reads until EOF without memory limits.
**Prevention:** Always initialize `&http.Client{Timeout: ...}` and wrap response bodies in `io.LimitReader(resp.Body, maxBytes)` before reading.
