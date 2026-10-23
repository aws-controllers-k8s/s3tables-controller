<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-091: A policy Deny on Put/DeleteTableBucketPolicy for the caller is enforced on the first call (AccessDeniedException 403...
_Full entry and notes of one finding; its summary entry is in
[tablebucket-policy.md](../tablebucket-policy.md). Generated from aws-api-semantics `services/s3tables` (model
2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-091"></a>**S3T-TABLEBUCKET-091** `delete-semantics` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **A policy Deny on Put/DeleteTableBucketPolicy for the caller is enforced on the first call (AccessDeniedException 403...**
  Put(Deny PutTableBucketPolicy+DeleteTableBucketPolicy for the caller role) -> HTTP 200. Re-Put attempts at 1
  s: AccessDeniedException 403 (0 accepted before the first 403; first 403 0.03 s after the Deny Put
  returned). DeleteTableBucketPolicy -> AccessDeniedException (HTTP 403) 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform:
  s3tables:DeleteTableBucketPolicy on resource: arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-ffbb82-polv
  with an explicit deny in a resource-based policy'. GetTableBucketPolicy -> HTTP 200 (version v1 in place).
  DeleteTableBucket (empty bucket) -> HTTP 204. Recreating the same name 3 s later -> ConflictException 409
  - ACK: terminal_codes, pre-delete-cleanup · ops: PutTableBucketPolicy, DeleteTableBucketPolicy,
    GetTableBucketPolicy, DeleteTableBucket · fields: resourcePolicy
  - repro: Put a Deny on Put/DeleteTableBucketPolicy for the caller; retry Put (same Deny + marker) at 1 s
    until 403; Delete; Get; DeleteTableBucket; recreate; Get
  - measurements: policy_enforcement_lag_s=0.03, puts_accepted_before_403=0, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-045](../tablebucket-policy.md#s3t-tablebucket-045), [S3T-TABLEBUCKET-046](../tablebucket-policy.md#s3t-tablebucket-046), [S3T-TABLEBUCKET-129](../tablebucket-policy.md#s3t-tablebucket-129), [S3T-TABLEBUCKETPOLICY-005](../tablebucket-policy.md#s3t-tablebucketpolicy-005),
    [S3T-TABLEBUCKETPOLICY-006](../tablebucket-policy.md#s3t-tablebucketpolicy-006), [S3T-TABLEBUCKETPOLICY-007](../tablebucket-policy.md#s3t-tablebucketpolicy-007), [S3T-TABLEBUCKETPOLICY-016](../tablebucket-policy.md#s3t-tablebucketpolicy-016), [S3T-TABLEPOLICY-009](../table-policy.md#s3t-tablepolicy-009) ·
    hypotheses: H-S-141 · evidence: tablebucket/weird-inputs/policy-validation-lockout

## Notes

H-S-141 qualified: lockout confirmed and DeleteTableBucket (204) is the escape, but the wire code is
AccessDeniedException 403 ('... with an explicit deny in a resource-based policy'), which the model does NOT
list for PutTableBucketPolicy/DeleteTableBucketPolicy (they list ForbiddenException). No enforcement lag: the
first Put 0.03 s after the Deny Put returned was already denied (n=1). CreateTableBucket with the same name 3
s after the delete -> ConflictException 409 'The bucket is in a transitional state because of a previous
deletion attempt. Try again later.' (the recreate step was not re-tried).

Full original title: A policy Deny on Put/DeleteTableBucketPolicy for the caller is enforced on the first call
(AccessDeniedException 403, not ForbiddenException); DeleteTableBucket escapes
