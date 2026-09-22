# Sentinel Security Journal

## 2025-05-18 - Missing Authentication on Process Cancellation Endpoint
**Vulnerability:** The `/cancel` endpoint allowed unauthenticated clients to send HTTP requests to terminate running process execution on the server.
**Learning:** Handlers added to `http.ServeMux` for execution management endpoints need explicit session checking (`GetSessionUser`) and HTTP method restrictions to prevent unauthorized access and Denial of Service.
**Prevention:** Ensure all state-modifying endpoints in web servers validate session authentication and HTTP methods before execution.
