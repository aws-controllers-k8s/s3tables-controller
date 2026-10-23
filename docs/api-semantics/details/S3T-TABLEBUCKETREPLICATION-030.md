<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKETREPLICATION-030: Deleted destination re-created under the same name 28.6 s later -> ConflictException 409; config re-pointed to a new bucket -> 200
_Full entry and notes of one finding; its summary entry is in
[tablebucket-replication.md](../tablebucket-replication.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucketreplication-030"></a>**S3T-TABLEBUCKETREPLICATION-030** `delete-semantics` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Deleted destination re-created under the same name 28.6 s later -> ConflictException 409; config re-pointed to a new bucket -> 200**
  CreateTableBucket(same destination name, other region) 28.6 s after its deletion -> ConflictException (HTTP 409)
  'The bucket is in a transitional state because of a previous deletion attempt. Try again later.'. Instead:
  new destination bucket + role policy, PutTableBucketReplication(new destination, current token) -> HTTP 200;
  the table that had failed/pended against the dead destination: [(0.0, 'failed'), (63.3, 'completed')]; the
  old 'completed' table afterwards 'pending'.
  - ACK: requeue, references · ops: CreateTableBucket, PutTableBucketReplication, GetTableReplicationStatus ·
    fields: configuration.rules[].destinations[].destinationTableBucketARN
  - repro: DeleteTableBucket(destination) -> CreateTableBucket(same name) -> PutTableBucketReplication(new
    destination) -> poll statuses
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-115](../tablebucket.md#s3t-tablebucket-115), [S3T-TABLEREPLICATION-017](../table-replication-status.md#s3t-tablereplication-017), [S3T-TABLEBUCKETREPLICATION-027](../tablebucket-replication.md#s3t-tablebucketreplication-027),
    [S3T-TABLEBUCKETREPLICATION-013](../tablebucket-replication.md#s3t-tablebucketreplication-013), [S3T-TABLEREPLICATION-008](../table-replication-status.md#s3t-tablereplication-008), [S3T-TABLEBUCKETREPLICATION-028](../tablebucket-replication.md#s3t-tablebucketreplication-028),
    [S3T-TABLEBUCKETREPLICATION-002](../tablebucket-replication.md#s3t-tablebucketreplication-002), [S3T-TABLEBUCKETREPLICATION-029](../tablebucket-replication.md#s3t-tablebucketreplication-029), [S3T-TABLEREPLICATION-005](../table-replication-status.md#s3t-tablereplication-005) · evidence:
    tablebucketreplication/creative/destination-deleted

## Notes

Cross-links the 10-min 'transitional state' name block ([S3T-TABLEBUCKET-115](../tablebucket.md#s3t-tablebucket-115)) with replication recovery: a
controller that re-creates a deleted destination TableBucket under the same name is blocked by 409 for ~10
min, while re-pointing the config to a new destination recovers at once. Single run.

Full original title: Re-creating the deleted destination under the same name 28.6 s later -> ConflictException
(HTTP 409); switching the config to a fresh destination -> HTTP 200, pending tables 'completed' at 63.3 s
