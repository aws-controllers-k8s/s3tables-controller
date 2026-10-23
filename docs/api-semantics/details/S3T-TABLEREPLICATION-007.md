<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEREPLICATION-007: DeleteTableReplication: table-token -> 409 'does not match the table version token', current -> 204, repeat -> 404; Get/Status -> 404...
_Full entry and notes of one finding; its summary entry is in [table-replication.md](../table-replication.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-tablereplication-007"></a>**S3T-TABLEREPLICATION-007** `delete-semantics` · impact high · unhandled (not handled in controller) · verified 2026-10-15, re-verified
  **DeleteTableReplication: table-token -> 409 'does not match the table version token', current -> 204, repeat -> 404; Get/Status -> 404...**
  DeleteTableReplication(versionToken = the table's own CreateTable/GetTable token) -> ConflictException 409
  'Provided version token does not match the table version token.' (a different message than Put's 'The
  specified version token does not match the configuration.'). DeleteTableReplication(current replication
  token) -> 204; GetTableReplication -> NotFoundException 404 'The specified replication configuration does
  not exist.'; GetTableReplicationStatus -> the same 404; repeating the Delete -> the same 404. The source
  table itself: GetTable -> 200, unchanged. (In this run the replica had already been deleted in the previous
  step, so what remains in the destination after a table-level delete is recorded in the bucket-level probe.)
  DeleteTable(source) afterwards -> 204.
  - ACK: custom_delete, pre-delete-cleanup, exceptions.404 · ops: DeleteTableReplication, GetTableReplication,
    GetTableReplicationStatus, UpdateTableMetadataLocation · fields: versionToken
  - repro: completed replication -> DeleteTableReplication(stale) -> DeleteTableReplication(current) ->
    Get/Status -> ListTables(destination) -> commit on source -> GetTable(replica) after 120 s
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-026](../tablebucket-replication.md#s3t-tablebucketreplication-026), [S3T-TABLEBUCKETREPLICATION-006](../tablebucket-replication.md#s3t-tablebucketreplication-006), [S3T-TABLEBUCKETREPLICATION-004](../tablebucket-replication.md#s3t-tablebucketreplication-004),
    [S3T-TABLEBUCKETREPLICATION-008](../tablebucket-replication.md#s3t-tablebucketreplication-008), [S3T-TABLEREPLICATION-036](../table-replication.md#s3t-tablereplication-036), [S3T-TABLEREPLICATION-015](../table-replication.md#s3t-tablereplication-015),
    [S3T-TABLEREPLICATION-001](../table-replication.md#s3t-tablereplication-001), [S3T-TABLEREPLICATION-010](../table-replication.md#s3t-tablereplication-010), [S3T-TABLEBUCKETREPLICATION-005](../tablebucket-replication.md#s3t-tablebucketreplication-005),
    [S3T-TABLEBUCKETREPLICATION-015](../tablebucket-replication.md#s3t-tablebucketreplication-015), [S3T-TABLEBUCKETREPLICATION-013](../tablebucket-replication.md#s3t-tablebucketreplication-013), [S3T-TABLEREPLICATION-003](../table-replication.md#s3t-tablereplication-003),
    [S3T-TABLEBUCKETREPLICATION-017](../tablebucket-replication.md#s3t-tablebucketreplication-017), [S3T-TABLEREPLICATION-004](../table-replication.md#s3t-tablereplication-004), [S3T-TABLEBUCKETREPLICATION-027](../tablebucket-replication.md#s3t-tablebucketreplication-027),
    [S3T-TABLEBUCKETREPLICATION-016](../tablebucket-replication.md#s3t-tablebucketreplication-016), [S3T-TABLEBUCKETREPLICATION-032](../tablebucket-replication.md#s3t-tablebucketreplication-032), [S3T-TABLE-138](../table-replication-status.md#s3t-table-138), [S3T-TABLE-196](../table-replication-status.md#s3t-table-196),
    [S3T-TABLEREPLICATION-044](../table-replication.md#s3t-tablereplication-044), [S3T-TABLEREPLICATION-016](../table-replication.md#s3t-tablereplication-016), [S3T-TABLEREPLICATION-018](../table-replication.md#s3t-tablereplication-018) · hypotheses: H-S-061,
    H-U-137 · evidence: table/round-trip/repl-put-get

## Notes

Token kinds are enforced on Delete too (H-U-137). H-S-061/H-S-087 at table level could not be judged here
because the replica was gone; see tablebucket/sub-resources/repl-backfill-delete for the bucket-level answer
(replicas remain, read-only).

Full original title: DeleteTableReplication: table-token -> 409 'does not match the table version token',
current -> 204, repeat -> 404; Get/Status -> 404; source GetTable unaffected
