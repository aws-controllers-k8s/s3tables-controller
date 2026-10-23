<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEREPLICATION-014: PutTableReplication response status is the literal 'success' (first Put, re-Put, dest add), not a ReplicationStatus; Get read-your-writes
_Full entry and notes of one finding; its summary entry is in [table-replication.md](../table-replication.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-tablereplication-014"></a>**S3T-TABLEREPLICATION-014** `shape-mismatch` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **PutTableReplication response status is the literal 'success' (first Put, re-Put, dest add), not a ReplicationStatus; Get read-your-writes** (hypothesis refuted; behavior confirmed)
  PutTableReplication response status: first Put -> 'success'; identical re-Put with the current token ->
  'success'; Put adding a second destination -> 'success'. None is in the GetTableReplicationStatus enum
  ['pending', 'completed', 'failed']. GetTableReplication x10 within ~1 s after the Put returned 1 distinct
  token(s), all equal to the Put's token.
  - ACK: synced.when, custom_update · ops: PutTableReplication, GetTableReplication · fields: status,
    versionToken
  - repro: Put -> Get x10 @100 ms -> re-Put identical (token) -> Put [D, E] (token); compare status strings
  - measurements: get_x10_distinct_tokens=[1], n=10
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEREPLICATION-003](../table-replication.md#s3t-tablereplication-003), [S3T-TABLEREPLICATION-038](../table-replication.md#s3t-tablereplication-038), [S3T-TABLEREPLICATION-039](../table-replication.md#s3t-tablereplication-039),
    [S3T-TABLEREPLICATION-044](../table-replication.md#s3t-tablereplication-044), [S3T-TABLEREPLICATION-029](../table-replication-status.md#s3t-tablereplication-029), [S3T-TABLEBUCKETREPLICATION-002](../tablebucket-replication.md#s3t-tablebucketreplication-002) · hypotheses: H-U-139,
    H-U-160 · evidence: table/dependencies/repl-prereqs-token-matrix

## Notes

H-U-139 REFUTED: the Put status is a literal operation result ('success'), not the per-destination enum, so it
cannot seed status conditions. H-U-160 first half (read-your-writes on GetTableReplication) confirmed (10/10).

Full original title: PutTableReplication.status is 'success' on first Put, identical re-Put and destination
add - not a ReplicationStatus value; Get x10 read-your-writes
