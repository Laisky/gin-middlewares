This candidate preserves github.com/Laisky/gin-middlewares/v5 and original v5 dependency PR histories #4,#6,#7. JWT remains on v4 at4.5.2 for GO-2025-3553. Patched x/net v0.60.0 covers current HTTP/2 advisories; x/crypto is0.57.0. The compiler requirement is Go1.26.0, with qualification on patched Go1.26.9. Full authentication/metrics race tests, module verification, vet and build pass.

Your code is affected by 0 vulnerabilities.
This scan also found 1 vulnerability in packages you import and 1 vulnerability
in modules you require, but your code doesn't appear to call these
vulnerabilities.
Use '-show verbose' for more details.

No finding is suppressed. Await exact-head Snyk. This v5 candidate is not retargeted to v7.
