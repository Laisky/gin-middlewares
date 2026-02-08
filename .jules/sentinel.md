## 2026-02-08 - [Panic in GetUserClaims]
**Vulnerability:** Denial of Service (DoS) via malformed Authorization header.
**Learning:** Slicing a string based on prefix length without checking the actual string length can cause a panic.
**Prevention:** Use safer string manipulation functions like `strings.TrimPrefix` and `strings.TrimSpace` or perform explicit length checks.
