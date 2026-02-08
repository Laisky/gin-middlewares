## 2026-02-08 - [Cookie Options Silent Failure]
**Vulnerability:** The `SetCookie` function accepted variadic options (`opts ...SetCookieOption`) but failed to pass them to the underlying `applyOpts` call. This caused security flags like `HttpOnly` and `Secure` to be silently ignored even when explicitly requested by the developer.
**Learning:** Variadic parameters in Go must be explicitly passed with the `...` syntax (e.g., `applyOpts(opts...)`). Forgetting this leads to the function receiving an empty slice instead of the intended options.
**Prevention:** Always verify that security-related options are correctly applied in unit tests. Added a permanent `cookie_test.go` to ensure these flags are correctly propagated to the browser.
