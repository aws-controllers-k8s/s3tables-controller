<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKETREPLICATION-014: Table created 485 s after the bucket-level Put (warm-up over): status 'completed' in 4.13 s (n=1); replica present in both destinations
_Full entry and notes of one finding; its summary entry is in
[tablebucket-replication.md](../tablebucket-replication.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucketreplication-014"></a>**S3T-TABLEBUCKETREPLICATION-014** `async-state-machine` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Table created 485 s after the bucket-level Put (warm-up over): status 'completed' in 4.13 s (n=1); replica present in both destinations**
  CreateTable (no metadata) in the source bucket 485.2s after PutTableBucketReplication:
  GetTableReplicationStatus transitions 0.05s -> no-destinations; 2.09s -> pending|pending; 4.13s ->
  completed|completed; terminal after 4.13s. Destination D then lists ['ackq_19dd08_b1', 'ackq_19dd08_b2',
  'ackq_19dd08_b3', 'ackq_19dd08_b4', 'ackq_19dd08_b5', 'ackq_19dd08_b6', 'ackq_a5e7ff_t3']; same-region
  destination E lists ['ackq_19dd08_b1', 'ackq_19dd08_b2', 'ackq_19dd08_b3', 'ackq_19dd08_b4',
  'ackq_19dd08_b5', 'ackq_19dd08_b6'].
  - ACK: e2e-timing, post-create-nudge · ops: CreateTable, GetTableReplicationStatus, ListTables · fields:
    destinations[].replicationStatus
  - repro: PutTableBucketReplication -> CreateTable -> poll GetTableReplicationStatus(new table)
  - measurements: new_table_to_terminal_s=[4.13], n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-027](../tablebucket-replication.md#s3t-tablebucketreplication-027), [S3T-TABLEBUCKETREPLICATION-028](../tablebucket-replication.md#s3t-tablebucketreplication-028), [S3T-TABLEBUCKETREPLICATION-031](../tablebucket-replication.md#s3t-tablebucketreplication-031),
    [S3T-TABLEBUCKETREPLICATION-032](../tablebucket-replication.md#s3t-tablebucketreplication-032), [S3T-TABLEREPLICATION-021](../table-replication-status.md#s3t-tablereplication-021), [S3T-TABLEREPLICATION-022](../table-replication-status.md#s3t-tablereplication-022),
    [S3T-TABLEREPLICATION-018](../table-replication.md#s3t-tablereplication-018), [S3T-TABLEBUCKETREPLICATION-013](../tablebucket-replication.md#s3t-tablebucketreplication-013), [S3T-TABLEBUCKETREPLICATION-015](../tablebucket-replication.md#s3t-tablebucketreplication-015) · hypotheses:
    H-S-062, H-S-058 · evidence: tablebucket/sub-resources/repl-backfill-delete

## Notes

H-S-062 second half confirmed: tables created after the Put are replicated.

Contradiction with [S3T-TABLEBUCKETREPLICATION-027](../tablebucket-replication.md#s3t-tablebucketreplication-027): 014 (n=1, table created 485.2 s after the bucket-level Put)
titles 'Table created after the bucket-level Put: status completed after 4.13s'; 027's t1 created 1 s after
the Put stayed 'none' until +61.4 s and completed at +62.4 s (031: 66.0 s, 032: 61.0 s, 029: 33.3 s). 031
reconciles by time since the Put: the ~60 s warm-up belongs to the configuration, not to each table - tables
created after it complete in 2-5 s (027 3.2 s, 028 2.1-3.2 s, 031/032 5.2 s). Resolution: keep both;
[S3T-TABLEBUCKETREPLICATION-031](../tablebucket-replication.md#s3t-tablebucketreplication-031) is canonical for the reconciliation (time since Put); 014 over-generalizes in
its title, see title_fixes
