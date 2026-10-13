# go-utils v6.4.0 consumer update

The v7 middleware module retains its major and updates go-utils/v6 from the
qualified 506ec9758d5f pseudo-version to v6.4.0 (6a02887d3dbf60ecadf21a67b5d3b9449fd942e6).
Used APIs are JWT signing/parsing, tracing, context-aware logging, time and
format helpers. The consumer does not call PasswordHash, legacy crypto,
CopyFile or memory Search, so no storage-format adaptation is introduced.

The retained middleware tests cover missing/malformed bearer headers, invalid
signature/expiry, context logger propagation, trace-ID handling, request
redaction, panic safety and canceled-request background work. Run the same
race suite against the original and updated graphs. Local and exact-head CI
receipts are recorded in the Draft PR description; publication alone is not
acceptance.

Adopt through a rebuilt canary after CI review and keep the old dependency
commit/binary for rollback. No sessions, policies or stored data are migrated.
This task does not merge or deploy.

This repository previously had no test workflow. The new read-only
go-utils-consumer workflow checks the immutable original consumer graph and
candidate with Go1.26.9 and1.27.2. It retains all race tests and candidate
vet/build/module-consistency checks, uses no secrets or live services, and
cannot publish or deploy. Snyk settings are unchanged.
