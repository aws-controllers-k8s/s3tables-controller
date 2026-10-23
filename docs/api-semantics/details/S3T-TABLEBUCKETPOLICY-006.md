<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKETPOLICY-006: Caller-scoped Deny on Get/Put/DeleteTableBucketPolicy locks the policy...
_Full entry and notes of one finding; its summary entry is in
[tablebucket-policy.md](../tablebucket-policy.md). Generated from aws-api-semantics `services/s3tables` (model
2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucketpolicy-006"></a>**S3T-TABLEBUCKETPOLICY-006** `delete-semantics` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Caller-scoped Deny on Get/Put/DeleteTableBucketPolicy locks the policy...**
  PutTableBucketPolicy(Deny Principal=<caller session ARN>, Action Get/Put/DeleteTableBucketPolicy,
  Resource=<bucket ARN>) -> HTTP 200. 1 s later, as the same caller: GetTableBucketPolicy ->
  AccessDeniedException (HTTP 403) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to
  perform: s3tables:GetTableBucketPolicy on resource:
  arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-a0e652-fc-51-tags with an explicit deny in a resource-based
  policy'; PutTableBucketPolicy -> AccessDeniedException (HTTP 403) 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform: s3tables:PutTableBucketPolicy
  on resource: arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-a0e652-fc-51-tags with an explicit deny in a
  resource-based policy'; DeleteTableBucketPolicy -> AccessDeniedException (HTTP 403) 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform:
  s3tables:DeleteTableBucketPolicy on resource:
  arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-a0e652-fc-51-tags with an explicit deny in a resource-based
  policy'; GetTableBucket -> HTTP 200; ListTagsForResource -> HTTP 200; GetTableBucketEncryption -> HTTP 200;
  ListNamespaces -> HTTP 200; DeleteTableBucket (empty bucket) -> HTTP 204; GetTableBucket afterwards ->
  NotFoundException/404.
  - ACK: pre-delete-cleanup, terminal_codes · ops: PutTableBucketPolicy, GetTableBucketPolicy,
    DeleteTableBucketPolicy, DeleteTableBucket
  - repro: PutTableBucketPolicy with a Deny for the caller on the three policy ops -> try each policy op, the
    bucket reads and DeleteTableBucket
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-091](../tablebucket-policy.md#s3t-tablebucket-091), [S3T-TABLEBUCKET-045](../tablebucket-policy.md#s3t-tablebucket-045), [S3T-TABLEBUCKET-046](../tablebucket-policy.md#s3t-tablebucket-046), [S3T-TABLEBUCKET-129](../tablebucket-policy.md#s3t-tablebucket-129),
    [S3T-TABLEBUCKETPOLICY-005](../tablebucket-policy.md#s3t-tablebucketpolicy-005), [S3T-TABLEBUCKETPOLICY-007](../tablebucket-policy.md#s3t-tablebucketpolicy-007), [S3T-TABLEBUCKETPOLICY-016](../tablebucket-policy.md#s3t-tablebucketpolicy-016) · evidence:
    tablebucket/creative/failed-create-name-reservation

## Notes

Extends [S3T-TABLEBUCKET-091](../tablebucket-policy.md#s3t-tablebucket-091) (policy Deny enforced on Put/Delete) with the ReadOne side (GetTableBucketPolicy)
and the recovery path: a TableBucketPolicy CR under such a policy cannot be read, updated or deleted by the
controller (AccessDeniedException on all three), while the parent bucket's reads and DeleteTableBucket are
unaffected, so deleting the TableBucket is the only way to clear the finalizer. The Deny was scoped to the
caller's session ARN so the state is recoverable with another session. Single run.

Full original title: Caller-scoped Deny on Get/Put/DeleteTableBucketPolicy locks the policy
(AccessDeniedException/AccessDeniedException/AccessDeniedException) but DeleteTableBucket still -> HTTP 204
