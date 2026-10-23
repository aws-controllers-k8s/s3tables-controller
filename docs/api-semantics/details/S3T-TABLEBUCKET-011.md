<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-011: PutTableBucketReplication accepts a role ARN that does not exist (200, status 'success'); a 5 s-old role was accepted on the first try
_Full entry and notes of one finding; its summary entry is in
[tablebucket-replication.md](../tablebucket-replication.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-011"></a>**S3T-TABLEBUCKET-011** `prerequisite` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **PutTableBucketReplication accepts a role ARN that does not exist (200, status 'success'); a 5 s-old role was accepted on the first try** (hypothesis refuted; behavior confirmed)
  PutTableBucketReplication(configuration.role = ARN of a role name that does not exist in the account, one
  existing destination bucket in another region) -> HTTP 200 {versionToken, status: 'success'};
  GetTableBucketReplication then returned that configuration (role is not validated at Put time). With the
  shard role created by iam:CreateRole (trust replication.s3tables.amazonaws.com) the first Put issued 4.9 s
  after CreateRole returned 200 on attempt 1 (no IAM-propagation rejection observed, n=1). The source bucket
  was EMPTY (no namespace, no table) and the Put still returned 200.
  - ACK: requeue, references · ops: PutTableBucketReplication, CreateRole · fields: configuration.role
  - repro: iam:CreateRole (trust replication.s3tables.amazonaws.com) -> immediately PutTableBucketReplication
    on an empty bucket, retry every 2 s on 4xx
  - measurements: first_put_ok_since_role_create_s=[4.9], rejected_attempts_before_ok=[0], n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-021](../tablebucket-replication.md#s3t-tablebucketreplication-021), [S3T-TABLEBUCKETREPLICATION-009](../tablebucket-replication.md#s3t-tablebucketreplication-009), [S3T-TABLEBUCKETREPLICATION-017](../tablebucket-replication.md#s3t-tablebucketreplication-017),
    [S3T-TABLEREPLICATION-032](../table-replication-status.md#s3t-tablereplication-032), [S3T-TABLEREPLICATION-002](../table-replication-status.md#s3t-tablereplication-002), [S3T-TABLEREPLICATION-008](../table-replication-status.md#s3t-tablereplication-008), [S3T-TABLEREPLICATION-028](../table-replication-status.md#s3t-tablereplication-028),
    [S3T-TABLEBUCKET-021](../tablebucket-replication.md#s3t-tablebucket-021), [S3T-TABLEBUCKETREPLICATION-003](../tablebucket-replication.md#s3t-tablebucketreplication-003), [S3T-SERVICE-021](../service.md#s3t-service-021) · hypotheses: H-S-054, H-S-062 ·
    evidence: tablebucket/dependencies/repl-setup

## Notes

H-S-054 REFUTED: a nonexistent role is NOT rejected with 400 (accepted with 200); consequently no
IAM-propagation 4xx can be seen on this path (n=1 role, first Put at +4.9 s). H-S-062 confirmed: Put on an
empty source bucket succeeds. The role-existence check, if any, is asynchronous (see the table-level status
probes).
