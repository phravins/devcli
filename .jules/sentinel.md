## 2025-05-18 - Local Web Server Origin Validation and Host Header Checks
**Vulnerability:** Cross-Site Request Forgery (CSRF) and DNS Rebinding vulnerability in the local web compiler server (`internal/web/server.go`). Cross-origin requests from browser pages or DNS rebinding could execute arbitrary code or shell commands via `/run` and `/terminal`.
**Learning:** Checking `Origin` or `Referer` headers using simple string prefix matching (`strings.HasPrefix(origin, "http://localhost")`) allows domain prefix bypasses (such as `http://localhost.attacker.com`).
**Prevention:** Always parse the URL with `url.Parse` and strictly compare `u.Hostname()` against `127.0.0.1` and `localhost` when validating request origins for local web servers.
