## 2025-05-18 - Unhandled Password Hashing Errors in Web Auth Registration
**Vulnerability:** In `internal/web/auth.go`, `bcrypt.GenerateFromPassword` errors were ignored (`hashed, _ := ...`). If password hashing failed (e.g. input exceeding bcrypt's 72-byte max length), the server saved user records with empty password hashes.
**Learning:** Functions returning error values must always be checked before using their return values, especially in cryptographic or security contexts where failure can lead to unauthenticated or corrupted user accounts.
**Prevention:** Always handle errors from `bcrypt.GenerateFromPassword` and return an explicit `500 Internal Server Error` response instead of proceeding with blank or zero-value credentials.
