<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEREPLICATION-005: GetTableReplicationStatus: destinations [] for ~60 s, then pending; empty table completed by +138 s, schema table at +436 s...
_Full entry and notes of one finding; its summary entry is in
[table-replication-status.md](../table-replication-status.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablereplication-005"></a>**S3T-TABLEREPLICATION-005** `async-state-machine` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **GetTableReplicationStatus: destinations [] for ~60 s, then pending; empty table completed by +138 s, schema table at +436 s...**
  Right after PutTableReplication the status body is {sourceTableArn, destinations: []} (empty list, no state)
  and stays so for about a minute. Relative to each table's Put: t1 (CreateTable with an Iceberg schema)
  destinations[0] appeared as replicationStatus='pending' with destinationTableArn at +61.9 s (the replica
  table was listable in the destination bucket at about the same time, +65 s) and became 'completed' at +435.7
  s with lastSuccessfulReplicatedUpdate{metadataLocation (source file name), timestamp}. t3 (CreateTable
  WITHOUT metadata) was already 'completed' at +138 s (first observation after a gap; it was still [] at +63
  s). t2 (well-formed ARN of a destination bucket that does not exist): Put -> 200; destinations[0] became
  'failed' at +63.6 s with failureMessage 'The specified destination bucket does not exist.' and no
  destinationTableArn.
  - ACK: synced.when, requeue, e2e-timing · ops: GetTableReplicationStatus, PutTableReplication · fields:
    destinations[].replicationStatus, destinations[].destinationTableArn,
    destinations[].lastSuccessfulReplicatedUpdate, destinations[].failureMessage
  - repro: PutTableReplication on three tables (schema / no metadata / missing destination) ->
    GetTableReplicationStatus every 2 s for <= 10 min
  - measurements: status_populated_after_put_s=[61.9, 63.6], empty_table_completed_by_s=[138.2],
    schema_table_pending_to_completed_s=[435.7], missing_destination_failed_at_s=[63.6], n=3
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEREPLICATION-033](../table-replication-status.md#s3t-tablereplication-033), [S3T-TABLEREPLICATION-019](../table-replication-status.md#s3t-tablereplication-019), [S3T-TABLEREPLICATION-028](../table-replication-status.md#s3t-tablereplication-028),
    [S3T-TABLEREPLICATION-032](../table-replication-status.md#s3t-tablereplication-032), [S3T-TABLEREPLICATION-017](../table-replication-status.md#s3t-tablereplication-017), [S3T-TABLEBUCKETREPLICATION-009](../tablebucket-replication.md#s3t-tablebucketreplication-009),
    [S3T-TABLEBUCKETREPLICATION-017](../tablebucket-replication.md#s3t-tablebucketreplication-017), [S3T-TABLEBUCKETREPLICATION-011](../tablebucket-replication.md#s3t-tablebucketreplication-011), [S3T-TABLEBUCKETREPLICATION-010](../tablebucket-replication.md#s3t-tablebucketreplication-010),
    [S3T-TABLEBUCKETREPLICATION-029](../tablebucket-replication.md#s3t-tablebucketreplication-029), [S3T-TABLEBUCKETREPLICATION-013](../tablebucket-replication.md#s3t-tablebucketreplication-013), [S3T-TABLEBUCKETREPLICATION-027](../tablebucket-replication.md#s3t-tablebucketreplication-027),
    [S3T-TABLEREPLICATION-006](../table-replication-status.md#s3t-tablereplication-006), [S3T-TABLEBUCKETREPLICATION-030](../tablebucket-replication.md#s3t-tablebucketreplication-030), [S3T-TABLEBUCKET-115](../tablebucket.md#s3t-tablebucket-115), [S3T-TABLEREPLICATION-031](../table-replication-status.md#s3t-tablereplication-031),
    [S3T-TABLEREPLICATION-038](../table-replication.md#s3t-tablereplication-038), [S3T-TABLEREPLICATION-025](../table-replication-status.md#s3t-tablereplication-025), [S3T-TABLEREPLICATION-030](../table-replication-status.md#s3t-tablereplication-030), [S3T-TABLEREPLICATION-043](../table-replication-status.md#s3t-tablereplication-043) ·
    hypotheses: H-U-134, H-U-135, H-S-058, H-S-062 · evidence: table/round-trip/repl-put-get

## Notes

H-U-134 (synchronous 404 for a missing destination) REFUTED: Put returned 200. H-U-135 confirmed: the failure
is asynchronous, 'failed' with a message naming the missing bucket at ~64 s. H-S-058/H-S-062 are covered at
bucket level in tablebucket/sub-resources/repl-backfill-delete. Self-heal after creating the bucket later was
not tested. Timings are from one run (the watch was restarted once, timestamps re-derived from evidence).

Full original title: GetTableReplicationStatus: destinations [] for ~60 s, then pending; empty table completed
by +138 s, schema table at +436 s; missing destination failed at +64 s
