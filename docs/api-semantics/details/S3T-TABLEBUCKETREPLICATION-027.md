<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKETREPLICATION-027: Identical PutTableBucketReplication re-send (token rotates) does not re-replicate 'completed' tables (150 s) or restart a backfill
_Full entry and notes of one finding; its summary entry is in
[tablebucket-replication.md](../tablebucket-replication.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucketreplication-027"></a>**S3T-TABLEBUCKETREPLICATION-027** `idempotency` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Identical PutTableBucketReplication re-send (token rotates) does not re-replicate 'completed' tables (150 s) or restart a backfill**
  Bucket-level config (one other-region destination) put at T0; table t1 (no metadata) created at T0+1 s:
  status 'none' until +61.4 s, 'pending' +61.4 s, 'completed' +62.4 s; table t2 (schema) created at T0+2 s:
  'pending' from the first read, 'completed' only at ~+336 s (lastSuccessfulReplicatedUpdate.timestamp).
  Identical re-Put (GetTableBucketReplication.configuration == sent document) with the current token -> HTTP
  200, token rotated, response status 'success'. GetTableReplicationStatus polled at 1/s per table for 150 s
  after the re-Put: t1 stayed 'completed' (single value for 150 s) and its replica's GetTable (modifiedAt,
  versionToken) was byte-identical before/after; t2's in-flight backfill continued unchanged ('pending' ->
  'completed' at +95 s of the watch, i.e. not restarted). A table created 6 s after the re-Put reached
  'completed' in 3.2 s (no new ~60 s warm-up).
  - ACK: compare.is_ignored+delta_pre_compare, one-per-reconcile · ops: PutTableBucketReplication,
    GetTableBucketReplication, GetTableReplicationStatus · fields: versionToken, configuration,
    destinations[].replicationStatus
  - repro: PutTableBucketReplication -> CreateTable x2 -> wait 'completed' -> GetTableBucketReplication token
    -> identical Put -> poll GetTableReplicationStatus 150 s + GetTable(replica)
  - measurements: t1_completed_s=62.4, t2_completed_s=336, post_reput_watch_s=150,
    new_table_after_reput_completed_s=3.2, n=2
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-002](../tablebucket-replication.md#s3t-tablebucketreplication-002), [S3T-TABLEREPLICATION-031](../table-replication-status.md#s3t-tablereplication-031), [S3T-TABLEBUCKETREPLICATION-006](../tablebucket-replication.md#s3t-tablebucketreplication-006),
    [S3T-TABLEBUCKETREPLICATION-017](../tablebucket-replication.md#s3t-tablebucketreplication-017), [S3T-TABLEREPLICATION-004](../table-replication.md#s3t-tablereplication-004), [S3T-TABLEREPLICATION-007](../table-replication.md#s3t-tablereplication-007),
    [S3T-TABLEBUCKETREPLICATION-005](../tablebucket-replication.md#s3t-tablebucketreplication-005), [S3T-TABLEREPLICATION-010](../table-replication.md#s3t-tablereplication-010), [S3T-TABLEREPLICATION-005](../table-replication-status.md#s3t-tablereplication-005),
    [S3T-TABLEBUCKETREPLICATION-013](../tablebucket-replication.md#s3t-tablebucketreplication-013), [S3T-TABLEREPLICATION-006](../table-replication-status.md#s3t-tablereplication-006), [S3T-TABLEBUCKETREPLICATION-029](../tablebucket-replication.md#s3t-tablebucketreplication-029),
    [S3T-TABLEBUCKETREPLICATION-014](../tablebucket-replication.md#s3t-tablebucketreplication-014), [S3T-TABLEBUCKETREPLICATION-028](../tablebucket-replication.md#s3t-tablebucketreplication-028), [S3T-TABLEBUCKETREPLICATION-031](../tablebucket-replication.md#s3t-tablebucketreplication-031),
    [S3T-TABLEBUCKETREPLICATION-032](../tablebucket-replication.md#s3t-tablebucketreplication-032), [S3T-TABLEBUCKETREPLICATION-030](../tablebucket-replication.md#s3t-tablebucketreplication-030), [S3T-TABLEREPLICATION-008](../table-replication-status.md#s3t-tablereplication-008) · evidence:
    tablebucketreplication/creative/identical-reput-retrigger

## Notes

Extends [S3T-TABLEBUCKETREPLICATION-002](../tablebucket-replication.md#s3t-tablebucketreplication-002) (identical re-Put rotates the token) with its effect on replication:
none - a controller that re-Puts on every reconcile only churns the versionToken (invalidating any token a
user or a second controller holds) and does not cause backfills, so a perpetual spec diff is harmless to the
data but makes the token in status stale every reconcile. Also observed: the first table after a fresh Put
reports no destination for ~61 s (empty table completed at 62 s) while a schema table took ~336 s; tables
created later complete in 2-3 s. n=2 tables, one re-Put, single run.

Full original title: Identical PutTableBucketReplication re-send (token rotates) does NOT re-trigger
replication: 'completed' table stays completed 150 s, replica unchanged

Contradiction with [S3T-TABLEBUCKETREPLICATION-014](../tablebucket-replication.md#s3t-tablebucketreplication-014): 014 (n=1, table created 485.2 s after the bucket-level Put)
titles 'Table created after the bucket-level Put: status completed after 4.13s'; 027's t1 created 1 s after
the Put stayed 'none' until +61.4 s and completed at +62.4 s (031: 66.0 s, 032: 61.0 s, 029: 33.3 s). 031
reconciles by time since the Put: the ~60 s warm-up belongs to the configuration, not to each table - tables
created after it complete in 2-5 s (027 3.2 s, 028 2.1-3.2 s, 031/032 5.2 s). Resolution: keep both;
[S3T-TABLEBUCKETREPLICATION-031](../tablebucket-replication.md#s3t-tablebucketreplication-031) is canonical for the reconciliation (time since Put); 014 over-generalizes in
its title, see title_fixes

Contradiction with [S3T-TABLEREPLICATION-008](../table-replication-status.md#s3t-tablereplication-008): 027's (truncated) title says an identical re-Put 'does NOT
re-trigger replication' - observed only for 'completed' tables (unchanged 150 s) and an in-flight backfill
(not restarted) under a bucket-level rule; TR-008 observed an identical table-level re-Put with the current
token re-trigger a 'failed' configuration to 'completed' in 65.88 s after the trust policy was fixed (waiting
150 s did not). 030 adds that a CHANGED document re-pends a 'completed' table. The direction depends on the
per-destination state (completed/in-flight untouched, failed re-attempted) and on whether the document
changed. Resolution: keep both; state-dependent, neither record is wrong - 027 over-generalizes in its
truncated title, see title_fixes
