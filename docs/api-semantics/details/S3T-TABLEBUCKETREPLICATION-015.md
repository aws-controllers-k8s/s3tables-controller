<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKETREPLICATION-015: After DeleteTableBucketReplication: the 6 replicas remain (read-only, 400 on update), per-table status -> 404...
_Full entry and notes of one finding; its summary entry is in
[tablebucket-replication.md](../tablebucket-replication.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucketreplication-015"></a>**S3T-TABLEBUCKETREPLICATION-015** `delete-semantics` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **After DeleteTableBucketReplication: the 6 replicas remain (read-only, 400 on update), per-table status -> 404...**
  DeleteTableBucketReplication(current token) -> 204; GetTableBucketReplication -> NotFoundException 404;
  GetTableReplicationStatus(b1) -> NotFoundException 404 'The specified replication configuration does not
  exist.' (status is gone with the configuration); the source bucket GetTableBucket -> 200 unchanged. Both
  destinations still list all 6 replicas (b1-b6). The replica of b1 in D keeps
  managedByService='replication.s3tables.amazonaws.com' and managedTableInformation.replicationInformation;
  UpdateTableMetadataLocation on it -> BadRequestException 400 'Updates to service managed tables are not
  allowed.'. A new commit on source b1 and a new table b7 created after the Delete: after 120 s the replica's
  metadataLocation was unchanged and b7 did not appear in either destination; GetTableReplicationStatus(b7)
  and (b1) -> 404.
  - ACK: pre-delete-cleanup, custom_delete, docs-only · ops: DeleteTableBucketReplication,
    GetTableReplicationStatus, ListTables, UpdateTableMetadataLocation, GetTable · fields: managedByService,
    managedTableInformation.replicationInformation
  - repro: backfilled bucket config -> DeleteTableBucketReplication -> reads on source/destination -> commit +
    CreateTable on source -> wait 120 s -> reads
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-024](../tablebucket-replication.md#s3t-tablebucketreplication-024), [S3T-TABLEBUCKETREPLICATION-025](../tablebucket-replication.md#s3t-tablebucketreplication-025), [S3T-TABLEREPLICATION-035](../table-replication-status.md#s3t-tablereplication-035),
    [S3T-TABLEBUCKETREPLICATION-016](../tablebucket-replication.md#s3t-tablebucketreplication-016), [S3T-TABLEREPLICATION-018](../table-replication.md#s3t-tablereplication-018), [S3T-TABLE-225](../table-replication-status.md#s3t-table-225), [S3T-TABLEBUCKETREPLICATION-005](../tablebucket-replication.md#s3t-tablebucketreplication-005),
    [S3T-TABLEBUCKETREPLICATION-008](../tablebucket-replication.md#s3t-tablebucketreplication-008), [S3T-TABLEREPLICATION-001](../table-replication.md#s3t-tablereplication-001), [S3T-TABLEREPLICATION-010](../table-replication.md#s3t-tablereplication-010),
    [S3T-TABLEREPLICATION-007](../table-replication.md#s3t-tablereplication-007), [S3T-TABLEBUCKETREPLICATION-013](../tablebucket-replication.md#s3t-tablebucketreplication-013), [S3T-TABLEREPLICATION-003](../table-replication.md#s3t-tablereplication-003),
    [S3T-TABLEBUCKETREPLICATION-032](../tablebucket-replication.md#s3t-tablebucketreplication-032), [S3T-TABLE-138](../table-replication-status.md#s3t-table-138), [S3T-TABLE-196](../table-replication-status.md#s3t-table-196), [S3T-TABLEREPLICATION-021](../table-replication-status.md#s3t-tablereplication-021),
    [S3T-TABLEREPLICATION-022](../table-replication-status.md#s3t-tablereplication-022), [S3T-TABLEBUCKETREPLICATION-014](../tablebucket-replication.md#s3t-tablebucketreplication-014) · hypotheses: H-S-061, H-S-087 · evidence:
    tablebucket/sub-resources/repl-backfill-delete

## Notes

H-S-061 confirmed: replicas remain, stay read-only and stop following the source; the source bucket is
unaffected. H-S-087 REFUTED: the orphaned replicas are NOT released to writable tables
(UpdateTableMetadataLocation -> 400); they can only be read or deleted.

Full original title: After DeleteTableBucketReplication: the 6 replicas remain (read-only, 400 on update),
per-table status -> 404, later commits/new tables are not replicated
