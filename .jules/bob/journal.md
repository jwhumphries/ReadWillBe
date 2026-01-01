## 2024-03-24 - [GZIP Middleware Config]
**Learning:** `middleware.Gzip()` with default settings compresses even tiny responses where CPU cost outweighs bandwidth savings.
**Action:** Configure `Level: 5` and `MinLength: 1400` (MTU size) for better efficiency.
