## 2024-03-24 - [GZIP Middleware Config]
**Learning:** `middleware.Gzip()` with default settings compresses even tiny responses where CPU cost outweighs bandwidth savings.
**Action:** Configure `Level: 5` and `MinLength: 1400` (MTU size) for better efficiency.

## 2024-05-23 - Session Optimization and Middleware
**Learning:** Checking for data changes before calling `session.Save()` significantly reduces unnecessary I/O and `Set-Cookie` headers, especially for long-lived sessions. Adding `BodyLimit` and `RequestID` are low-effort high-value improvements for security and observability.
**Action:** Always verify if a session write is necessary before invoking the save operation in middleware. Always include basic protection middleware like BodyLimit in Echo servers.
