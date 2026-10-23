<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-SERVICE-012: Role without s3tables permissions: 15/15 never-denied ops -> AccessDeniedException 403; IAM allow visible after 86.1 s, revoke after 100.5 s
_Full entry and notes of one finding; its summary entry is in [service.md](../service.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-service-012"></a>**S3T-SERVICE-012** `error-code` · impact high · handled · verified 2026-10-15, re-verified
  **Role without s3tables permissions: 15/15 never-denied ops -> AccessDeniedException 403; IAM allow visible after 86.1 s, revoke after 100.5 s**
  Assumed role with no identity policy: AccessDeniedException/403 'User: <ARN> is not authorized to perform:
  s3tables:<Op> on resource: <ARN> because no identity-based policy allows the s3tables:<Op> action' on
  ['DeleteTableBucketEncryption', 'DeleteTableBucketMetricsConfiguration', 'DeleteTableBucketReplication',
  'DeleteTableReplication', 'GetTableBucket (control)', 'GetTableBucketMetricsConfiguration',
  'GetTableBucketReplication', 'GetTableBucketStorageClass', 'GetTableMaintenanceJobStatus',
  'GetTableReplication', 'GetTableReplicationStatus', 'PutTableBucketEncryption',
  'PutTableBucketMaintenanceConfiguration', 'PutTableBucketMetricsConfiguration',
  'PutTableBucketStorageClass']. Exceptions: none. Then PutRolePolicy Allow s3tables:GetTableBucket:
  GetTableBucket codes by second ['AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException',
  'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException',
  'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException',
  'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException',
  'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException',
  'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException',
  'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException',
  'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException',
  'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException',
  'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException',
  'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException',
  'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException',
  'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException',
  'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException',
  'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException',
  'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException',
  'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException',
  'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException',
  'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException',
  'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException',
  'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException',
  'AccessDeniedException', 'AccessDeniedException', 'OK'] (first 200 at 86.13 s); DeleteRolePolicy: codes
  ['OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK',
  'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK',
  'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK',
  'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK',
  'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK',
  'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'AccessDeniedException'] (first AccessDeniedException
  at 100.5 s).
  - ACK: terminal_codes, requeue · ops: GetTableBucketStorageClass, GetTableBucketReplication,
    GetTableBucketMetricsConfiguration, PutTableBucketMetricsConfiguration,
    DeleteTableBucketMetricsConfiguration, PutTableBucketEncryption, DeleteTableBucketEncryption,
    PutTableBucketStorageClass, PutTableBucketMaintenanceConfiguration, DeleteTableBucketReplication,
    GetTableReplication, GetTableReplicationStatus, DeleteTableReplication, GetTableMaintenanceJobStatus,
    GetTableBucket (control)
  - repro: CreateRole (no policy) -> AssumeRole -> call each op; PutRolePolicy Allow GetTableBucket and poll
    at 1 s; DeleteRolePolicy and poll at 1 s
  - measurements: allow_visible_s=86.13, revoke_visible_s=100.5, n=1
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-SERVICE-019](../service.md#s3t-service-019), [S3T-SERVICE-010](../service.md#s3t-service-010), [S3T-SERVICE-025](../service.md#s3t-service-025), [S3T-TABLEBUCKETPOLICY-001](../tablebucket-policy.md#s3t-tablebucketpolicy-001), [S3T-TABLE-164](../table-policy.md#s3t-table-164),
    [S3T-SERVICE-024](../service.md#s3t-service-024), [S3T-SERVICE-027](../service.md#s3t-service-027), [S3T-SERVICE-011](../service.md#s3t-service-011), [S3T-TABLE-234](../table-maintenance.md#s3t-table-234) · evidence:
    service/creative/iam-authz-variants

## Notes

Creative round, seed error-messages: AccessDeniedException is declared on only 15 of 49 operations in the
model yet is reachable on every one (the structured probes saw ~35; this adds the remaining ones listed in
operations). It is AUTHZ but can be transient for seconds after an IAM change (allow/revoke windows above,
n=1), so a controller should requeue with backoff rather than mark the resource terminal.
DeleteTableBucketReplication/DeleteTableReplication were sent with a garbage token to see whether authz or the
token check wins.

Full original title: Role without s3tables permissions: 15/15 never-before-denied ops -> AccessDeniedException
403; Allow visible after 86.13 s, revoke after 100.5 s
