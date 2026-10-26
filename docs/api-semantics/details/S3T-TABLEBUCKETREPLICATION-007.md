<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKETREPLICATION-007: PutTableBucketReplication returns status='success'; GetTableBucketReplication has only configuration+versionToken...
_Full entry and notes of one finding; its summary entry is in
[tablebucket-replication.md](../tablebucket-replication.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucketreplication-007"></a>**S3T-TABLEBUCKETREPLICATION-007** `read-gap` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **PutTableBucketReplication returns status='success'; GetTableBucketReplication has only configuration+versionToken...**
  PutTableBucketReplication response members: versionToken and status='success' (a literal operation result).
  GetTableBucketReplication at +0 s and at the end of the probe (several minutes later) returned exactly the
  members ['configuration', 'versionToken'] - no status, health or progress member exists in the bucket-level
  read path.
  - ACK: synced.when, docs-only · ops: PutTableBucketReplication, GetTableBucketReplication · fields: status,
    versionToken, configuration
  - repro: Put -> inspect response; Get at +0 and +5 min -> list top-level members
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-018](../tablebucket-replication.md#s3t-tablebucketreplication-018), [S3T-TABLEBUCKETREPLICATION-001](../tablebucket-replication.md#s3t-tablebucketreplication-001), [S3T-TABLEBUCKETREPLICATION-003](../tablebucket-replication.md#s3t-tablebucketreplication-003),
    [S3T-TABLEREPLICATION-037](../table-replication.md#s3t-tablereplication-037), [S3T-TABLEREPLICATION-003](../table-replication.md#s3t-tablereplication-003), [S3T-TABLEREPLICATION-001](../table-replication.md#s3t-tablereplication-001),
    [S3T-TABLEBUCKETREPLICATION-010](../tablebucket-replication.md#s3t-tablebucketreplication-010), [S3T-TABLEREPLICATION-009](../table-replication.md#s3t-tablereplication-009) · hypotheses: H-S-048 · evidence:
    tablebucket/dependencies/repl-setup

## Notes

H-S-048 qualified: the Put response status is the literal lowercase 'success' (the user guide shows
'Success'); it is an operation result, not a replication state, and the Get has no status member. Progress is
only visible per table via GetTableReplicationStatus (probed separately).

Full original title: PutTableBucketReplication returns status='success'; GetTableBucketReplication has only
configuration+versionToken - no replication status at bucket level
