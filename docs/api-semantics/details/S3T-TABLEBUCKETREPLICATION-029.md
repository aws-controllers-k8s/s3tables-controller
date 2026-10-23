<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKETREPLICATION-029: Destination bucket deleted under an active config: Get still echoes it, old table stays 'completed'; new table and source commit -> 'failed'
_Full entry and notes of one finding; its summary entry is in
[tablebucket-replication.md](../tablebucket-replication.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucketreplication-029"></a>**S3T-TABLEBUCKETREPLICATION-029** `async-state-machine` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Destination bucket deleted under an active config: Get still echoes it, old table stays 'completed'; new table and source commit -> 'failed'**
  Bucket-level config to one other-region destination; table t1 'completed' at 33.3 s. Destination emptied
  (replicas ['ackq_704fdb_t1'] deleted 204, namespace 204) and DeleteTableBucket -> HTTP 204. 2 s later
  GetTableBucketReplication -> HTTP 200 and still names the deleted destination: True; t1 status completed.
  New table t2 created: status transitions [(0.1, 'none'), (23.9, 'failed')] (final entry
  {"replicationStatus": "failed", "failureMessage": "The specified destination bucket does not exist."}). t1
  status afterwards 'completed'. Metadata-only UpdateTableMetadataLocation on t1 -> HTTP 200; its status over
  the next 90 s: [(0.0, 'pending'), (2.1, 'failed')].
  - ACK: references, requeue · ops: DeleteTableBucket, GetTableBucketReplication, GetTableReplicationStatus,
    CreateTable, UpdateTableMetadataLocation · fields:
    configuration.rules[].destinations[].destinationTableBucketARN, destinations[].replicationStatus,
    destinations[].failureMessage
  - repro: Bucket-level replication -> table completed -> empty+delete the destination bucket ->
    GetTableBucketReplication, GetTableReplicationStatus, CreateTable, commit on the old table
  - measurements: t1_completed_s=33.3, t2_terminal_s=23.9, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-009](../tablebucket-replication.md#s3t-tablebucketreplication-009), [S3T-TABLEREPLICATION-002](../table-replication-status.md#s3t-tablereplication-002), [S3T-TABLEBUCKETREPLICATION-017](../tablebucket-replication.md#s3t-tablebucketreplication-017),
    [S3T-TABLEBUCKETREPLICATION-011](../tablebucket-replication.md#s3t-tablebucketreplication-011), [S3T-TABLEBUCKETREPLICATION-010](../tablebucket-replication.md#s3t-tablebucketreplication-010), [S3T-TABLEREPLICATION-005](../table-replication-status.md#s3t-tablereplication-005),
    [S3T-TABLEBUCKETREPLICATION-013](../tablebucket-replication.md#s3t-tablebucketreplication-013), [S3T-TABLEBUCKETREPLICATION-027](../tablebucket-replication.md#s3t-tablebucketreplication-027), [S3T-TABLEREPLICATION-006](../table-replication-status.md#s3t-tablereplication-006),
    [S3T-TABLEBUCKETREPLICATION-030](../tablebucket-replication.md#s3t-tablebucketreplication-030), [S3T-TABLEBUCKET-115](../tablebucket.md#s3t-tablebucket-115) · evidence:
    tablebucketreplication/creative/destination-deleted

## Notes

Extends [S3T-TABLEBUCKETREPLICATION-009](../tablebucket-replication.md#s3t-tablebucketreplication-009) (Put validates nothing) with the post-hoc destruction of the
destination: the configuration read-back never reflects a dead destination and already-replicated tables keep
reporting 'completed', so a TableBucketReplication controller cannot detect the broken state from its own
resource; only per-table statuses of new tables/commits show it. Single run.

Full original title: Destination bucket deleted under an active config: config still echoes it (200), old
status stays 'completed', new table -> 'failed' at 23.9 s, source commit -> 'failed'
