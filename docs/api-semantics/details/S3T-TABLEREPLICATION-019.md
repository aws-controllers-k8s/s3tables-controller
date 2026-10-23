<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEREPLICATION-019: Empty table, table-level Put: status [] until +61.06s, pending until +62.09s, then 'completed'; replica listable in D at +62.2s (n=1)
_Full entry and notes of one finding; its summary entry is in
[table-replication-status.md](../table-replication-status.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablereplication-019"></a>**S3T-TABLEREPLICATION-019** `async-state-machine` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Empty table, table-level Put: status [] until +61.06s, pending until +62.09s, then 'completed'; replica listable in D at +62.2s (n=1)**
  PutTableReplication(empty table -> D) -> HTTP 200. GetTableReplicationStatus at 1 s: [(0.06,
  'no-destinations'), (61.06, 'pending'), (62.09, 'completed')]. The replica table (same namespace/name) was
  first listed by ListTables(D) at +62.2s (polled every ~5 s). Raw entry while pending: {"destinations":
  [{"destinationTableBucketArn": "arn:aws:s3tables:us-east-1:<ACCOUNT>:bucket/ackq-b73e55-repl-dest",
  "replicationStatus": "pending"}], "sourceTableArn":
  "arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-f81240-lc-src/table/46c786d8-2add-407d-a37c-66608a1da538"}.
  Raw entry at terminal: {"destinations": [{"destinationTableArn":
  "arn:aws:s3tables:us-east-1:<ACCOUNT>:bucket/ackq-b73e55-repl-dest/table/93cf68ab-77f5-416b-b185-18cbf19578e8",
  "destinationTableBucketArn": "arn:aws:s3tables:us-east-1:<ACCOUNT>:bucket/ackq-b73e55-repl-dest",
  "replicationStatus": "completed"}], "sourceTableArn":
  "arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-f81240-lc-src/table/46c786d8-2add-407d-a37c-66608a1da538"}.
  Note the key spelling destinationTableBucketArn (status) vs destinationTableBucketARN (configuration).
  - ACK: synced.when, e2e-timing, requeue · ops: PutTableReplication, GetTableReplicationStatus, ListTables ·
    fields: destinations[].replicationStatus, destinations[].destinationTableArn,
    destinations[].destinationTableBucketArn, destinations[].lastSuccessfulReplicatedUpdate
  - repro: CreateTable(no metadata) -> PutTableReplication -> GetTableReplicationStatus @1 s + ListTables(D)
    @5 s until terminal
  - measurements: no_destinations_window_s=[61.06], pending_to_terminal_s=[1.03], put_to_terminal_s=[62.09],
    replica_listable_s=[62.2], n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEREPLICATION-005](../table-replication-status.md#s3t-tablereplication-005), [S3T-TABLEBUCKETREPLICATION-013](../tablebucket-replication.md#s3t-tablebucketreplication-013), [S3T-TABLEREPLICATION-033](../table-replication-status.md#s3t-tablereplication-033),
    [S3T-TABLEREPLICATION-028](../table-replication-status.md#s3t-tablereplication-028), [S3T-TABLEREPLICATION-032](../table-replication-status.md#s3t-tablereplication-032), [S3T-TABLEREPLICATION-017](../table-replication-status.md#s3t-tablereplication-017), [S3T-TABLEREPLICATION-031](../table-replication-status.md#s3t-tablereplication-031),
    [S3T-TABLEREPLICATION-038](../table-replication.md#s3t-tablereplication-038), [S3T-TABLEREPLICATION-025](../table-replication-status.md#s3t-tablereplication-025), [S3T-TABLEREPLICATION-030](../table-replication-status.md#s3t-tablereplication-030), [S3T-TABLEREPLICATION-043](../table-replication-status.md#s3t-tablereplication-043),
    [S3T-TABLEREPLICATION-041](../table-replication-status.md#s3t-tablereplication-041), [S3T-TABLE-136](../table-replication-status.md#s3t-table-136), [S3T-TABLE-224](../table-replication-status.md#s3t-table-224), [S3T-TABLEREPLICATION-029](../table-replication-status.md#s3t-tablereplication-029) · hypotheses: H-U-151,
    H-U-152, H-U-161, H-U-162, H-U-197 · evidence: table/state-machine/repl-status-lifecycle

## Notes

H-U-151 qualified: no 404 window, but the body is {sourceTableArn, destinations: []} (no entry at all) for ~1
min before 'pending'. H-U-152 REFUTED (empty table reached 'completed' without any commit). H-U-161 confirmed
/ H-U-162 refuted: namespace and table are created by the service in the empty destination bucket. H-U-197
confirmed: the entry carries destinationTableArn (replica ARN) only at terminal, so no ListTables is needed to
locate the replica.

Full original title: Empty table: status [] until +61.06s, pending until +62.09s, then 'completed'; replica
listable in D at +62.2s; entry members destinationTableArn+destinationTableBucketArn+replicationStatus

Contradiction with [S3T-TABLEREPLICATION-029](../table-replication-status.md#s3t-tablereplication-029): when destinationTableArn is present in a status entry: 019's
first-backfill 'pending' entry has only destinationTableBucketArn+replicationStatus and its notes conclude the
ARN is carried 'only at terminal'; 029's B2 'pending' after the corrective re-Put already carries
destinationTableArn (failureMessage gone) 455 s before 'completed'; 028 adds that 'failed' entries never carry
it; n=1 each Resolution: keep both; [S3T-TABLEREPLICATION-029](../table-replication-status.md#s3t-tablereplication-029) narrows 019: destinationTableArn is absent on
'failed' and on the first-backfill 'pending' but present on a re-Put 'pending' and at 'completed' - presence
is per entry and optional in every non-completed state, not a terminal marker
