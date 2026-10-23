<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-SERVICE-028: Re-verified: identity-policy Allow on an assumed role becomes visible to s3tables after 86.1/95.2 s, revoke after 95.4/95.3 s (n=2...
_Full entry and notes of one finding; its summary entry is in [service.md](../service.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-service-028"></a>**S3T-SERVICE-028** `eventual-consistency` · impact medium · unhandled (not handled in controller) · verified 2026-10-15
  **Re-verified: identity-policy Allow on an assumed role becomes visible to s3tables after 86.1/95.2 s, revoke after 95.4/95.3 s (n=2...**
  Role with no identity policy, one STS session: 15/15 ops -> AccessDeniedException/403 'User: <ARN> is not
  authorized to perform: s3tables:<Op> on resource: <ARN> because no identity-based policy allows the
  s3tables:<Op> action' (exceptions: none). PutRolePolicy(Allow s3tables:GetTableBucket) -> GetTableBucket
  polled at 1/s flipped from AccessDeniedException to 200 after [86.1, 95.2] s; DeleteRolePolicy -> flipped
  back to AccessDeniedException after [95.4, 95.3] s; both flips were stable for the next 5 s in every cycle
  (True / True).
  - ACK: requeue, e2e-timing · ops: GetTableBucket, DeleteTableBucketEncryption,
    DeleteTableBucketMetricsConfiguration, DeleteTableBucketReplication, DeleteTableReplication,
    GetTableBucketMetricsConfiguration, GetTableBucketReplication, GetTableBucketStorageClass,
    GetTableMaintenanceJobStatus, GetTableReplication, GetTableReplicationStatus, PutTableBucketEncryption,
    PutTableBucketMaintenanceConfiguration, PutTableBucketMetricsConfiguration, PutTableBucketStorageClass
  - repro: CreateRole(no policy) -> AssumeRole -> 15 ops; 2x (PutRolePolicy Allow GetTableBucket, poll 1/s;
    DeleteRolePolicy, poll 1/s)
  - measurements: allow_visible_s=[86.1, 95.2], revoke_visible_s=[95.4, 95.3], n=2
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-SERVICE-012](../service.md#s3t-service-012) · evidence: service/creative/reverify-set-e

## Notes

Second-opinion reproduction of [S3T-SERVICE-012](../service.md#s3t-service-012) (which had n=1: 86.1 s / 100.5 s). Same assumed-role session
reused across cycles, so the measured delay includes any per-session authorization caching.

Full original title: Re-verified: identity-policy Allow on an assumed role becomes visible to s3tables after
86.1/95.2 s, revoke after 95.4/95.3 s (n=2, same STS session)
