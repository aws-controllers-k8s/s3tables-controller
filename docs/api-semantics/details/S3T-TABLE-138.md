<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-138: DeleteTable(replica) with source config active -> 204; status stays 'completed' until the next source commit turns it 'failed'; no re-create
_Full entry and notes of one finding; its summary entry is in
[table-replication-status.md](../table-replication-status.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-138"></a>**S3T-TABLE-138** `delete-semantics` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **DeleteTable(replica) with source config active -> 204; status stays 'completed' until the next source commit turns it 'failed'; no re-create**
  DeleteTable on the service-created replica (destination region) while the source table's replication
  configuration is active -> HTTP 204 (replicas are deletable). GetTableReplicationStatus(source) still said
  'completed' right after and 60 s later; the replica did not come back on its own. After a new
  UpdateTableMetadataLocation commit on the source, ListTables(destination) polled every 5 s for 300 s never
  showed the replica again, and the status was then 'failed' with failureMessage 'The destination table has
  been deleted during replication.' (destinationTableArn and lastSuccessfulReplicatedUpdate still present).
  - ACK: pre-delete-cleanup, custom_delete · ops: DeleteTable, GetTableReplicationStatus,
    UpdateTableMetadataLocation · fields: destinations[].replicationStatus, destinations[].failureMessage
  - repro: replica completed -> DeleteTable(replica) -> status; sleep 60; commit on source -> poll
    ListTables(destination) 5 min
  - measurements: replica_not_recreated_within_s=[300], n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-139](../table-replication-status.md#s3t-table-139), [S3T-TABLE-196](../table-replication-status.md#s3t-table-196), [S3T-TABLEBUCKET-118](../tablebucket-replication.md#s3t-tablebucket-118), [S3T-TABLEBUCKET-012](../tablebucket-replication.md#s3t-tablebucket-012), [S3T-TABLE-306](../table-replication-status.md#s3t-table-306),
    [S3T-TABLEBUCKETREPLICATION-015](../tablebucket-replication.md#s3t-tablebucketreplication-015), [S3T-TABLEBUCKETREPLICATION-016](../tablebucket-replication.md#s3t-tablebucketreplication-016), [S3T-TABLEBUCKETREPLICATION-032](../tablebucket-replication.md#s3t-tablebucketreplication-032),
    [S3T-TABLEREPLICATION-007](../table-replication.md#s3t-tablereplication-007), [S3T-TABLEBUCKETREPLICATION-005](../tablebucket-replication.md#s3t-tablebucketreplication-005), [S3T-TABLEREPLICATION-025](../table-replication-status.md#s3t-tablereplication-025),
    [S3T-TABLEREPLICATION-020](../table-replication-status.md#s3t-tablereplication-020), [S3T-TABLEREPLICATION-022](../table-replication-status.md#s3t-tablereplication-022), [S3T-TABLEREPLICATION-027](../table-replication-status.md#s3t-tablereplication-027), [S3T-TABLEREPLICATION-043](../table-replication-status.md#s3t-tablereplication-043) ·
    hypotheses: H-S-060 · evidence: table/round-trip/repl-put-get

## Notes

H-S-060 second half qualified: the replica is deletable, is NOT re-created on its own, and a later source
commit does NOT re-create it either - it moves the status to a terminal 'failed'. Recovery presumably needs a
new Put (not tested). Deleting replicas while replication is configured is therefore destructive.

Full original title: DeleteTable(replica) while the source config is active -> 204; status stays 'completed';
after the next source commit it turns 'failed' and the replica is not re-created
