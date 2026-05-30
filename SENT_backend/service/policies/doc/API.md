# Policy Service API - Baseline / Approve Endpoints

## List Policies
GET /api/v1/policies?status={status}&category={category}
- Response: JSON array of Policy objects.
- Notes: `status=BASELINE` returns proposed baseline items.

## Approve single policy
PUT /api/v1/policies/:id/approve
- Roles: requires `approval_final` permission
- Response: { message: "Chính sách đã được kích hoạt" }

## Bulk approve policies
POST /api/v1/policies/bulk-approve
- Body: { ids: [1,2,3] }
- Roles: requires `approval_final` permission
- Response: { message: "Đã phê duyệt hàng loạt chính sách" }

## Approve all baseline for an asset
POST /api/v1/policies/approve-by-asset
- Body: { asset_hwid: "HWID123" }
- Roles: requires `approval_final` permission
- Response: { message: "Đã phê duyệt baseline cho máy" }

## Internal: Agent -> send Baseline proposals
POST /api/v1/policies/internal/baseline
- Body: { org_id: number, asset_hwid: string, category: string, values: [ { ...structured entry... } ] }
- Structured entries depend on category, examples:
  - SOFTWARE_HASH: { file_hash, software_name, publisher, version }
  - USB_DEVICE: { device_hash, device_name, vid, pid, serial_number }
  - PORT: { port, process_name }

Notes:
- Baseline proposals are stored with `approval_status = BASELINE` and `is_active = false` until promoted by Admin.
- Legacy `whitelist_items` are migrated into `policies` on service migration (idempotent).
