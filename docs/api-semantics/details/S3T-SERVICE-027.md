<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-SERVICE-027: IAM Allow seen by s3tables after 100.8/99.8 s, revoke after 101.8/102.0 s, same as s3 (n=2): IAM-wide propagation, not an s3tables cache
_Full entry and notes of one finding; its summary entry is in [service.md](../service.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-service-027"></a>**S3T-SERVICE-027** `error-code` · impact high · unhandled (not handled in controller) · verified 2026-10-14
  **IAM Allow seen by s3tables after 100.8/99.8 s, revoke after 101.8/102.0 s, same as s3 (n=2): IAM-wide propagation, not an s3tables cache**
  One assumed role, one inline policy granting s3tables:GetTableBucket and s3:ListAllMyBuckets, polled at 1 s.
  After PutRolePolicy the first 200: s3tables GetTableBucket at [100.8, 99.8] s, s3 ListBuckets at [100.8,
  99.8] s. After DeleteRolePolicy the first 403: s3tables at [101.8, 102.0] s, s3 at [101.8, 102.0] s. Codes
  while waiting: s3tables ['AccessDeniedException', 'OK'] / s3 ['AccessDenied', 'OK'] (allow),
  ['AccessDeniedException', 'OK'] / ['AccessDenied', 'OK'] (revoke).
  - ACK: requeue · ops: GetTableBucket
  - repro: CreateRole; AssumeRole; PutRolePolicy(Allow s3tables:GetTableBucket + s3:ListAllMyBuckets); poll
    both at 1 s; DeleteRolePolicy; poll both; x2
  - measurements: allow_s3tables_s=[100.8, 99.8], allow_s3_s=[100.8, 99.8], revoke_s3tables_s=[101.8, 102.0],
    revoke_s3_s=[101.8, 102.0], n=2
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-SERVICE-012](../service.md#s3t-service-012), [S3T-SERVICE-024](../service.md#s3t-service-024), [S3T-TABLE-234](../table-maintenance.md#s3t-table-234), [S3T-SERVICE-011](../service.md#s3t-service-011) · evidence:
    service/creative/iam-propagation-compare

## Notes

Creative round, seed error-messages: decides whether the AccessDeniedException row's '~90 s after an IAM
change' is a property of s3tables (authorization decision cache) or of IAM. Verdict: IAM-wide propagation (s3
lags as much). Controller rule stays: requeue AccessDeniedException with backoff for at least 2 min after an
IAM change.

Full original title: IAM Allow seen by s3tables after 100.8/99.8 s vs s3 after 100.8/99.8 s; revoke after
101.8/102.0 s vs 101.8/102.0 s (n=2): IAM-wide propagation (s3 lags as much)
