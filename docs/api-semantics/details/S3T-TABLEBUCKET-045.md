<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-045: IAM Deny (identity policy or table-bucket policy) surfaces as AccessDeniedException 403, not ForbiddenException...
_Full entry and notes of one finding; its summary entry is in
[tablebucket-policy.md](../tablebucket-policy.md). Generated from aws-api-semantics `services/s3tables` (model
2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-045"></a>**S3T-TABLEBUCKET-045** `error-code` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **IAM Deny (identity policy or table-bucket policy) surfaces as AccessDeniedException 403, not ForbiddenException...** (hypothesis refuted; behavior confirmed)
  As a role whose identity policy explicitly denies s3tables:GetTableBucketPolicy: GetTableBucketPolicy ->
  AccessDeniedException (HTTP 403) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to
  perform: s3tables:GetTableBucketPolicy on resource:
  arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-bd98cf-perm-src with an explicit deny in an identity-based
  policy'. After PutTableBucketPolicy (admin) with an explicit Deny of s3tables:GetTableBucketEncryption for
  that role: GetTableBucketEncryption as the role -> AccessDeniedException (HTTP 403) 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform:
  s3tables:GetTableBucketEncryption on resource:
  arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-bd98cf-perm-src with an explicit deny in a resource-based
  policy' (before the policy: HTTP 200; control GetTableBucketStorageClass as the role: HTTP 200; admin
  GetTableBucketEncryption with the Deny policy in place: HTTP 200).
  - ACK: terminal_codes, exceptions.404 · ops: GetTableBucketPolicy, GetTableBucketEncryption,
    PutTableBucketPolicy
  - repro: AssumeRole(role with identity Deny) -> GetTableBucketPolicy; PutTableBucketPolicy(Deny role on
    GetTableBucketEncryption) -> GetTableBucketEncryption as the role
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-046](../tablebucket-policy.md#s3t-tablebucket-046), [S3T-TABLEBUCKET-129](../tablebucket-policy.md#s3t-tablebucket-129), [S3T-TABLEBUCKET-091](../tablebucket-policy.md#s3t-tablebucket-091), [S3T-TABLEBUCKETPOLICY-005](../tablebucket-policy.md#s3t-tablebucketpolicy-005),
    [S3T-TABLEBUCKETPOLICY-006](../tablebucket-policy.md#s3t-tablebucketpolicy-006), [S3T-TABLEBUCKETPOLICY-007](../tablebucket-policy.md#s3t-tablebucketpolicy-007), [S3T-TABLEBUCKETPOLICY-016](../tablebucket-policy.md#s3t-tablebucketpolicy-016), [S3T-TABLE-109](../table-policy.md#s3t-table-109),
    [S3T-TABLE-250](../table-policy.md#s3t-table-250), [S3T-TABLEBUCKET-038](../tablebucket.md#s3t-tablebucket-038), [S3T-TABLE-026](../table-policy.md#s3t-table-026), [S3T-TABLEBUCKET-181](../tablebucket-encryption-storage.md#s3t-tablebucket-181), [S3T-TABLE-131](../table-policy.md#s3t-table-131) · hypotheses:
    H-S-018 · evidence: service/error-taxonomy/repl-permissions

## Notes

H-S-018 REFUTED: an explicit Deny in an identity-based policy AND an explicit Deny in the table-bucket
(resource-based) policy both surface as AccessDeniedException 403 with the IAM-style message ('... with an
explicit deny in an identity-based policy' / '... in a resource-based policy'). ForbiddenException 403
'Unauthorized' was only seen for ARNs of another account ([S3T-TABLEBUCKET-021](../tablebucket-replication.md#s3t-tablebucket-021)). So the two 403 shapes split by
cause, not by operation: AccessDeniedException = policy evaluation says no (incl. PassRole);
ForbiddenException = resource not visible to the caller.

Full original title: IAM Deny (identity policy or table-bucket policy) surfaces as AccessDeniedException 403,
not ForbiddenException; Forbidden is for foreign/unknown ARNs
