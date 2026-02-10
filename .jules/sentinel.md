## 2026-02-10 - [Security Enhancement] Secure by Default for Cookies
**Vulnerability:** Cookies were defaulting to `HttpOnly: false`, making them accessible to client-side scripts and vulnerable to XSS-based theft.
**Learning:** The project's documentation and memory emphasized a "security-first approach," but the implementation was lagging behind in default values.
**Prevention:** Always verify that security defaults in the code match the stated architectural goals. Prefer "Secure by Default" (e.g., `HttpOnly: true`) and require explicit opt-out for less secure configurations.
