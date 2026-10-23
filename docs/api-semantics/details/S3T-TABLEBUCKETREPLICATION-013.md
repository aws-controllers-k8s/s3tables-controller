<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKETREPLICATION-013: Bucket-level backfill: 5 existing tables terminal in 71 s (empty) and 443-485 s (schema); replicas listable at ~+71 s...
_Full entry and notes of one finding; its summary entry is in
[tablebucket-replication.md](../tablebucket-replication.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucketreplication-013"></a>**S3T-TABLEBUCKETREPLICATION-013** `async-state-machine` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Bucket-level backfill: 5 existing tables terminal in 71 s (empty) and 443-485 s (schema); replicas listable at ~+71 s...**
  PutTableBucketReplication(B -> D in another region) on a bucket with 5 existing tables (b1-b3 created with
  an Iceberg schema, b4-b5 without metadata) -> 200. Under the bucket-level config GetTableReplication(b1) ->
  NotFoundException 404 'The specified replication configuration does not exist.' (no table-level config
  exists) while GetTableReplicationStatus(b1) works (destinations: [] at +0, populated at ~+61 s). A second
  Put at +5.2 s adding a same-region destination E with the current token -> 200 and Get showed both
  destinations (no 'backfill in progress' conflict). Per-table status polled every 2 s: b4/b5 (no metadata)
  completed at +70.7 s for both destinations; b1/b2 completed for D at +443.2 s, b3 at +484.5 s; E's entries
  completed between +436 s and +484 s. The same-region destination E listed all five replica tables at +70.7 s
  (created with the namespace at the start of the backfill, long before 'completed').
  - ACK: e2e-timing, synced.when, custom_update · ops: PutTableBucketReplication, GetTableReplicationStatus,
    GetTableReplication, ListTables · fields: destinations[].replicationStatus,
    destinations[].destinationTableArn
  - repro: 5 tables -> PutTableBucketReplication -> +5 s Put with an added destination -> poll per-table
    status 2 s for <= 10 min; ListTables on both destinations
  - measurements: existing_table_terminal_s_schema=[443.2, 443.2, 484.5],
    existing_table_terminal_s_empty=[70.7, 70.7], replicas_listable_in_dest_s=[70.7], n=5
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-008](../tablebucket-replication.md#s3t-tablebucketreplication-008), [S3T-TABLEREPLICATION-001](../table-replication.md#s3t-tablereplication-001), [S3T-TABLEREPLICATION-010](../table-replication.md#s3t-tablereplication-010),
    [S3T-TABLEBUCKETREPLICATION-005](../tablebucket-replication.md#s3t-tablebucketreplication-005), [S3T-TABLEREPLICATION-007](../table-replication.md#s3t-tablereplication-007), [S3T-TABLEBUCKETREPLICATION-015](../tablebucket-replication.md#s3t-tablebucketreplication-015),
    [S3T-TABLEREPLICATION-003](../table-replication.md#s3t-tablereplication-003), [S3T-TABLEREPLICATION-005](../table-replication-status.md#s3t-tablereplication-005), [S3T-TABLEBUCKETREPLICATION-027](../tablebucket-replication.md#s3t-tablebucketreplication-027),
    [S3T-TABLEREPLICATION-006](../table-replication-status.md#s3t-tablereplication-006), [S3T-TABLEBUCKETREPLICATION-029](../tablebucket-replication.md#s3t-tablebucketreplication-029), [S3T-TABLEBUCKETREPLICATION-030](../tablebucket-replication.md#s3t-tablebucketreplication-030),
    [S3T-TABLEREPLICATION-008](../table-replication-status.md#s3t-tablereplication-008), [S3T-TABLEBUCKETREPLICATION-028](../tablebucket-replication.md#s3t-tablebucketreplication-028), [S3T-TABLEBUCKETREPLICATION-002](../tablebucket-replication.md#s3t-tablebucketreplication-002),
    [S3T-TABLEREPLICATION-019](../table-replication-status.md#s3t-tablereplication-019), [S3T-TABLEREPLICATION-031](../table-replication-status.md#s3t-tablereplication-031), [S3T-TABLEREPLICATION-038](../table-replication.md#s3t-tablereplication-038), [S3T-TABLEREPLICATION-025](../table-replication-status.md#s3t-tablereplication-025),
    [S3T-TABLEREPLICATION-030](../table-replication-status.md#s3t-tablereplication-030), [S3T-TABLEREPLICATION-043](../table-replication-status.md#s3t-tablereplication-043), [S3T-TABLEREPLICATION-021](../table-replication-status.md#s3t-tablereplication-021), [S3T-TABLEREPLICATION-022](../table-replication-status.md#s3t-tablereplication-022),
    [S3T-TABLEREPLICATION-018](../table-replication.md#s3t-tablereplication-018), [S3T-TABLEBUCKETREPLICATION-014](../tablebucket-replication.md#s3t-tablebucketreplication-014) · hypotheses: H-S-058, H-S-063 · evidence:
    tablebucket/sub-resources/repl-backfill-delete

## Notes

H-S-058 confirmed for both same-region (E) and cross-region (D) destinations: existing tables are backfilled;
tables with an Iceberg metadata file take ~7-8 min, tables without any metadata ~70 s. H-S-063 confirmed: a
Put during the initial backfill is accepted (200, new token). Note the D ListTables in this run also contained
another probe's replicas, so only E's listing is used for the 'first listed' timing.

Full original title: Bucket-level backfill: 5 existing tables terminal in 71 s (empty) and 443-485 s (schema);
replicas listable at ~+71 s; Put during backfill -> 200
