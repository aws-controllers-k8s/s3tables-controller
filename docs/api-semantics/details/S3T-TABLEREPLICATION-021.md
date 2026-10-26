<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEREPLICATION-021: Table under bucket-level replication: GetTableReplication 404, Status 200; table-level Put (no token) -> 200, replaces bucket dest in status
_Full entry and notes of one finding; its summary entry is in
[table-replication-status.md](../table-replication-status.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablereplication-021"></a>**S3T-TABLEREPLICATION-021** `requested-vs-effective` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Table under bucket-level replication: GetTableReplication 404, Status 200; table-level Put (no token) -> 200, replaces bucket dest in status**
  PutTableBucketReplication(S -> D) -> HTTP 200; t1 (existing empty table) status transitions [(0.08,
  'no-destinations'), (1.12, 'D:pending'), (2.16, 'D:completed')]. Under bucket-level only:
  GetTableReplication(t1) -> NotFoundException (HTTP 404) 'The specified replication configuration does not
  exist.'; GetTableReplicationStatus(t1) -> HTTP 200. PutTableReplication(t1 -> E, no token) -> HTTP 200.
  GetTableReplication(t1) then -> 200 {"role": "arn:aws:iam::<ACCOUNT>:role/<PRINCIPAL>", "rules":
  [{"destinations": [{"destinationTableBucketARN":
  "arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-04fb03-bt-same"}]}]}; GetTableBucketReplication(S)
  unchanged -> 200 dests ['D']. Status transitions after the table-level Put: [(0.03, 'D:completed'), (61.04,
  'D:completed|E:pending'), (62.07, 'E:completed')]. ListTables(E, ns) afterwards: ['ackq_04fb03_t1'];
  ListTables(D, ns): ['ackq_04fb03_t1'].
  - ACK: custom_find, synced.when, compare.is_ignored+delta_pre_compare · ops: PutTableBucketReplication,
    GetTableReplication, GetTableReplicationStatus, PutTableReplication, ListTables · fields:
    configuration.rules[].destinations, destinations[].destinationTableBucketArn,
    destinations[].replicationStatus
  - repro: PutTableBucketReplication(S->D) -> status(t1) terminal -> GetTableReplication(t1) ->
    PutTableReplication(t1->E) -> poll status -> ListTables(E), ListTables(D)
  - measurements: bucket_level_t1_terminal_s=[2.16], table_level_e_terminal_s=[62.07], n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-013](../tablebucket-replication.md#s3t-tablebucketreplication-013), [S3T-TABLEBUCKETREPLICATION-023](../tablebucket-replication.md#s3t-tablebucketreplication-023), [S3T-TABLEBUCKETREPLICATION-008](../tablebucket-replication.md#s3t-tablebucketreplication-008),
    [S3T-TABLEBUCKETREPLICATION-005](../tablebucket-replication.md#s3t-tablebucketreplication-005), [S3T-TABLEREPLICATION-034](../table-replication.md#s3t-tablereplication-034), [S3T-TABLEREPLICATION-001](../table-replication.md#s3t-tablereplication-001),
    [S3T-TABLEREPLICATION-037](../table-replication.md#s3t-tablereplication-037), [S3T-TABLEREPLICATION-015](../table-replication.md#s3t-tablereplication-015), [S3T-TABLEREPLICATION-044](../table-replication.md#s3t-tablereplication-044), [S3T-TABLEREPLICATION-020](../table-replication-status.md#s3t-tablereplication-020),
    [S3T-TABLEREPLICATION-017](../table-replication-status.md#s3t-tablereplication-017), [S3T-TABLEREPLICATION-042](../table-replication-status.md#s3t-tablereplication-042), [S3T-TABLEREPLICATION-022](../table-replication-status.md#s3t-tablereplication-022), [S3T-TABLEREPLICATION-018](../table-replication.md#s3t-tablereplication-018),
    [S3T-TABLEBUCKETREPLICATION-014](../tablebucket-replication.md#s3t-tablebucketreplication-014), [S3T-TABLEBUCKETREPLICATION-015](../tablebucket-replication.md#s3t-tablebucketreplication-015), [S3T-TABLEREPLICATION-047](../table-replication-status.md#s3t-tablereplication-047) · hypotheses:
    H-U-179, H-U-180, H-U-181 · evidence: table/sub-resources/repl-bucket-vs-table-config

## Notes

H-U-179 confirmed (GetTableReplication 404 under bucket-level only, GetTableReplicationStatus 200 with the
bucket-level destination). H-U-181 REFUTED (table-level Put accepted with HTTP 200, no 409). H-U-180 confirmed
for the empty table: after the table-level Put [E] the status showed 'D:completed|E:pending' for ~1 s and then
only 'E:completed' - the bucket-level destination D was dropped from the table's status (table-level takes
precedence, not a union); the bucket-level replica already in D remained listed.

Contradiction with [S3T-TABLEREPLICATION-047](../table-replication-status.md#s3t-tablereplication-047), [S3T-TABLEREPLICATION-017](../table-replication-status.md#s3t-tablereplication-017), [S3T-TABLEREPLICATION-042](../table-replication-status.md#s3t-tablereplication-042),
[S3T-TABLEREPLICATION-022](../table-replication-status.md#s3t-tablereplication-022): FU2-04 settled by follow-up probe
tablereplication/state-machine/fu2-04-put-offset-tick: 042/021/022 hold; 017's +1.06 s did not reproduce. The
delay is a fixed ~61 s per Put, not a periodic tick: Puts issued 0.9, 5, 20, 40, 55 and 70 s after the last
observed status change (67-565 s after the first Put) all showed the new destination as 'pending' 61.0-64.1 s
later and 'completed' 1.0-1.1 s after that (empty table, same-region buckets); the first Put behaved
identically (pendi Resolution: keep all; [S3T-TABLEREPLICATION-047](../table-replication-status.md#s3t-tablereplication-047) is canonical for the settled rule
