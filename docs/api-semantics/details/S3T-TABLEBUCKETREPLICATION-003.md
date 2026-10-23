<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKETREPLICATION-003: GetTableBucketReplication echoes destinations in the order sent and the role ARN with path byte-identical...
_Full entry and notes of one finding; its summary entry is in
[tablebucket-replication.md](../tablebucket-replication.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucketreplication-003"></a>**S3T-TABLEBUCKETREPLICATION-003** `normalization` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **GetTableBucketReplication echoes destinations in the order sent and the role ARN with path byte-identical...**
  Put with destinations [S2 (us-west-2 bucket), D (us-east-1 bucket)] then Get -> order preserved; Put [D, S2]
  then Get -> order preserved (no sorting). configuration.role, a role ARN with a path segment
  (role/ackq/<name>), came back byte-identical. Put with configuration.role='S3TablesReplicationRole' (bare
  role name, 23 chars) -> BadRequestException (HTTP 400) with the MISLEADING message 'Invalid input: Must
  provide either table arn alone OR all of (table bucket arn, namespace, table name) together'.
  - ACK: compare.is_ignored+delta_pre_compare, references · ops: PutTableBucketReplication,
    GetTableBucketReplication · fields: configuration.rules[].destinations, configuration.role
  - repro: Put [B, A] -> Get; Put [A, B] -> Get; Put role='S3TablesReplicationRole'
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-018](../tablebucket-replication.md#s3t-tablebucketreplication-018), [S3T-TABLEBUCKETREPLICATION-001](../tablebucket-replication.md#s3t-tablebucketreplication-001), [S3T-TABLEBUCKETREPLICATION-007](../tablebucket-replication.md#s3t-tablebucketreplication-007),
    [S3T-TABLEREPLICATION-037](../table-replication.md#s3t-tablereplication-037), [S3T-TABLEREPLICATION-003](../table-replication.md#s3t-tablereplication-003), [S3T-TABLEREPLICATION-001](../table-replication.md#s3t-tablereplication-001),
    [S3T-TABLEBUCKETREPLICATION-019](../tablebucket-replication.md#s3t-tablebucketreplication-019), [S3T-TABLEBUCKETREPLICATION-020](../tablebucket-replication.md#s3t-tablebucketreplication-020), [S3T-TABLEBUCKETREPLICATION-002](../tablebucket-replication.md#s3t-tablebucketreplication-002),
    [S3T-TABLEBUCKETREPLICATION-017](../tablebucket-replication.md#s3t-tablebucketreplication-017), [S3T-TABLEBUCKETREPLICATION-004](../tablebucket-replication.md#s3t-tablebucketreplication-004), [S3T-TABLEBUCKETREPLICATION-005](../tablebucket-replication.md#s3t-tablebucketreplication-005),
    [S3T-TABLEBUCKETREPLICATION-022](../tablebucket-replication.md#s3t-tablebucketreplication-022), [S3T-TABLEBUCKETREPLICATION-009](../tablebucket-replication.md#s3t-tablebucketreplication-009), [S3T-TABLEBUCKETREPLICATION-011](../tablebucket-replication.md#s3t-tablebucketreplication-011),
    [S3T-TABLEREPLICATION-017](../table-replication-status.md#s3t-tablereplication-017), [S3T-SERVICE-026](../service.md#s3t-service-026), [S3T-SERVICE-007](../service.md#s3t-service-007), [S3T-SERVICE-015](../service.md#s3t-service-015), [S3T-SERVICE-021](../service.md#s3t-service-021),
    [S3T-TABLEBUCKET-011](../tablebucket-replication.md#s3t-tablebucket-011), [S3T-TABLEBUCKET-021](../tablebucket-replication.md#s3t-tablebucket-021), [S3T-TABLEREPLICATION-011](../table-replication.md#s3t-tablereplication-011), [S3T-TABLEREPLICATION-012](../table-replication.md#s3t-tablereplication-012),
    [S3T-TABLEREPLICATION-013](../table-replication.md#s3t-tablereplication-013) · hypotheses: H-S-116, H-S-066 · evidence: tablebucket/dependencies/repl-setup

## Notes

H-S-116 confirmed (order preserved, path kept, no rewriting). H-S-066 confirmed (bare name -> 400), with the
surprise that the 400 message talks about table identifiers, not about the role.

Full original title: GetTableBucketReplication echoes destinations in the order sent and the role ARN with
path byte-identical; bare role name -> BadRequestException
