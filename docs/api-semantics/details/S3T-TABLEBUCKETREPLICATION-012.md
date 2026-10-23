<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKETREPLICATION-012: Caller with s3tables:* but Deny iam:PassRole: PutTableBucketReplication and PutTableReplication -> AccessDeniedException 403 (iam:PassRole)
_Full entry and notes of one finding; its summary entry is in
[tablebucket-replication.md](../tablebucket-replication.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucketreplication-012"></a>**S3T-TABLEBUCKETREPLICATION-012** `error-code` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Caller with s3tables:* but Deny iam:PassRole: PutTableBucketReplication and PutTableReplication -> AccessDeniedException 403 (iam:PassRole)**
  Assuming a role whose identity policy allows s3tables:* and explicitly denies iam:PassRole:
  PutTableBucketReplication(role=that same role) -> AccessDeniedException (HTTP 403) 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform: iam:PassRole on resource:
  arn:aws:iam::<ACCOUNT>:role/<PRINCIPAL> with an explicit deny in an identity-based policy'.
  PutTableReplication(same configuration) -> AccessDeniedException (HTTP 403) 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform: iam:PassRole on resource:
  arn:aws:iam::<ACCOUNT>:role/<PRINCIPAL> with an explicit deny in an identity-based policy'. Control reads by
  the same caller: GetTableBucket -> HTTP 200, GetTableBucketReplication -> NotFoundException (HTTP 404) 'The
  specified replication configuration does not exist.'.
  - ACK: terminal_codes, docs-only · ops: PutTableBucketReplication, PutTableReplication · fields:
    configuration.role
  - repro: AssumeRole(role with Deny iam:PassRole) -> PutTableBucketReplication / PutTableReplication
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEREPLICATION-002](../table-replication-status.md#s3t-tablereplication-002), [S3T-TABLEREPLICATION-008](../table-replication-status.md#s3t-tablereplication-008), [S3T-TABLEBUCKETREPLICATION-017](../tablebucket-replication.md#s3t-tablebucketreplication-017),
    [S3T-TABLEBUCKETREPLICATION-028](../tablebucket-replication.md#s3t-tablebucketreplication-028), [S3T-TABLEREPLICATION-009](../table-replication.md#s3t-tablereplication-009), [S3T-TABLEREPLICATION-023](../table-replication.md#s3t-tablereplication-023),
    [S3T-TABLEREPLICATION-024](../table-replication.md#s3t-tablereplication-024), [S3T-TABLEREPLICATION-013](../table-replication.md#s3t-tablereplication-013), [S3T-TABLEREPLICATION-028](../table-replication-status.md#s3t-tablereplication-028), [S3T-TABLE-026](../table-policy.md#s3t-table-026) · hypotheses:
    H-S-052, H-U-133 · evidence: service/error-taxonomy/repl-permissions

## Notes

H-S-052 confirmed; H-U-133 PassRole half confirmed: both Put operations return AccessDeniedException 403 whose
message is the IAM-style 'User: <caller> is not authorized to perform: iam:PassRole on resource: <role> with
an explicit deny in an identity-based policy'.

Full original title: Caller with s3tables:* but Deny iam:PassRole: PutTableBucketReplication ->
AccessDeniedException 403, PutTableReplication -> AccessDeniedException 403
