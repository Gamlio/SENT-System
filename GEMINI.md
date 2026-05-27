# SENT-System Project Context & Rules

## 1. Tech Stack Context
- **Frontend**: React, Axios API Client, WebSockets.
- **Backend**: Golang (Gin Framework), Microservices architecture.
- **Database**: PostgreSQL (GORM) for core entities, MongoDB for high-frequency Telemetry Logs.

## 2. Coding Best Practices & Guardrails
- **Multi-Tenant Isolation**: Every database query MUST include `org_id` to prevent IDOR vulnerabilities.
- **Type Safety**: Avoid unsafe type assertions from Gin context. Always use type-safe extractors like `c.GetUint("org_id")`.
- **Database Rules**: When using `ON CONFLICT` constraints in GORM, verify that the corresponding model in `models.go` has a matching compound `uniqueIndex`.
- **Maker-Checker Workflow**: Do not insert dummy data into core tables for PENDING actions; instead, serialize the payload into the `SnapshotData` field of an `ApprovalTicket`.

## 3. Exclusion Rules
- Ignore `node_modules/`, `.git/`, `dist/`, `build/`, and `*.exe` files during review.