<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKETPOLICY-007: Deny+NotAction bucket/table policy for the account root principal is accepted (200) and bricks the bucket...
_Full entry and notes of one finding; its summary entry is in
[tablebucket-policy.md](../tablebucket-policy.md). Generated from aws-api-semantics `services/s3tables` (model
2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucketpolicy-007"></a>**S3T-TABLEBUCKETPOLICY-007** `delete-semantics` · impact high · unhandled (not handled in controller) · verified 2026-10-14
  **Deny+NotAction bucket/table policy for the account root principal is accepted (200) and bricks the bucket...**
  PutTableBucketPolicy({Effect: Deny, Principal: {AWS: <account root ARN>}, NotAction:
  s3tables:GetTableBucket, Resource: <bucket ARN>}) -> HTTP 200 (the analogous PutTablePolicy on the table ->
  200). Afterwards, for the owning account's IAM principal: still allowed ['GetTableBucket'];
  AccessDeniedException 403 'with an explicit deny in a resource-based policy' for ['DeleteNamespace',
  'DeleteTable', 'DeleteTableBucket', 'DeleteTableBucketPolicy', 'DeleteTablePolicy', 'GetNamespace',
  'GetTable(tableArn)', 'GetTable(triple)', 'GetTableBucketEncryption',
  'GetTableBucketMaintenanceConfiguration', 'GetTableBucketPolicy', 'GetTablePolicy', 'ListNamespaces',
  'ListTables', 'ListTagsForResource(bucket)', 'PutTableBucketPolicy(open)', 'TagResource(bucket)']. The
  bucket is still listed by ListTableBuckets (['ackq-9ba36e-polnorm']). The account root principal in a
  resource-policy Deny matches every IAM principal of the account, so no IAM principal can delete the policy,
  the table, the namespace or the bucket; only the root user (or AWS Support) can recover.
  - ACK: pre-delete-cleanup, terminal_codes, is_iam_policy · ops: PutTableBucketPolicy, PutTablePolicy,
    DeleteTableBucketPolicy, DeleteTablePolicy, DeleteTable, DeleteNamespace, DeleteTableBucket · fields:
    resourcePolicy
  - repro: PutTableBucketPolicy(Deny, Principal root, NotAction s3tables:GetTableBucket, Resource bucket ARN)
    -> any other call on the bucket as any IAM principal of the account
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-214](../table-policy.md#s3t-table-214), [S3T-TABLEBUCKET-091](../tablebucket-policy.md#s3t-tablebucket-091), [S3T-TABLE-157](../table-policy.md#s3t-table-157), [S3T-TABLEBUCKET-045](../tablebucket-policy.md#s3t-tablebucket-045), [S3T-TABLEBUCKET-046](../tablebucket-policy.md#s3t-tablebucket-046),
    [S3T-TABLEBUCKET-129](../tablebucket-policy.md#s3t-tablebucket-129), [S3T-TABLEBUCKETPOLICY-005](../tablebucket-policy.md#s3t-tablebucketpolicy-005), [S3T-TABLEBUCKETPOLICY-006](../tablebucket-policy.md#s3t-tablebucketpolicy-006), [S3T-TABLEBUCKETPOLICY-016](../tablebucket-policy.md#s3t-tablebucketpolicy-016),
    [S3T-TABLEPOLICY-009](../table-policy.md#s3t-tablepolicy-009) · evidence: tablebucketpolicy/creative/lockout-recovery

## Notes

Extends [S3T-TABLE-214](../table-policy.md#s3t-table-214) / [S3T-TABLEBUCKET-091](../tablebucket-policy.md#s3t-tablebucket-091) (caller-scoped lockouts, recoverable by deleting the parent) with
the unrecoverable case: a Deny that also covers Delete* on the parent. PutTableBucketPolicy rejects public
Allow '*' policies but accepts an account-wide Deny-NotAction without any safety check. A
TableBucketPolicy/TablePolicy CR applying a user-supplied document can therefore make the parent
TableBucket/Table CRs undeletable (finalizers hang on 403 forever); controllers should at least surface the
403 'explicit deny in a resource-based policy' as a terminal condition instead of requeueing silently. The
locked lab bucket ackq-9ba36e-polnorm (namespace ackq_9ba36e_ns, table ackq_9ba36e_t) stays registered and
needs root credentials to remove.

Full original title: Deny+NotAction bucket/table policy for the account root principal is accepted (200) and
bricks the bucket: Delete* -> AccessDeniedException, no owner override
