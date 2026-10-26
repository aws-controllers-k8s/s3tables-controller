<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEREPLICATION-022: Table-level Put [D,E] where D holds the bucket-level replica -> D 'failed' (identical name exists); table-level delete -> bucket view
_Full entry and notes of one finding; its summary entry is in
[table-replication-status.md](../table-replication-status.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablereplication-022"></a>**S3T-TABLEREPLICATION-022** `requested-vs-effective` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Table-level Put [D,E] where D holds the bucket-level replica -> D 'failed' (identical name exists); table-level delete -> bucket view**
  CreateTable t2 while S has bucket-level replication to D (and t1 has a table-level config to E): status(t2)
  transitions [(0.06, 'no-destinations'), (1.1, 'D:pending'), (2.13, 'D:completed')]; GetTableReplication(t2)
  -> NotFoundException (HTTP 404) 'The specified replication configuration does not exist.'; ListTables(E) =
  ['ackq_04fb03_t1'] (t2 not sent to the table-level destination of t1). Table-level Put [D, E] on t1 (D also
  a bucket-level destination) -> HTTP 200; status transitions [(0.06, 'E:completed'), (61.17,
  'D:pending|E:completed'), (62.2, 'D:failed|E:completed')]; status body {"destinations":
  [{"destinationTableBucketArn": "arn:aws:s3tables:us-east-1:<ACCOUNT>:bucket/ackq-b73e55-repl-dest",
  "failureMessage": "A table with an identical name already exists at the destination.", "replicationStatus":
  "failed"}, {"destinationTableArn":
  "arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-04fb03-bt-same/table/eb9e8561-0ee2-4507-a849-c3a42cf6d0b0",
  "destinationTableBucketArn": "arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-04fb03-bt-same",
  "replicationStatus": "completed"}], "sourceTableArn":
  "arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-04fb03-bt-src/table/0e1ad48c-783d-4b4a-aecc-4bebbd6e9103"}.
  DeleteTableReplication(t1) -> HTTP 204; status(t1) over 60 s: [(0.1, 'D:failed|E:completed'), (61.3,
  'D:failed')]; GetTableReplication(t1) -> NotFoundException (HTTP 404) 'The specified replication
  configuration does not exist.'. DeleteTableBucketReplication(S) -> HTTP 204; status(t1)/status(t2) over 30
  s: [(0.1, 't1=ERR:NotFoundException t2=ERR:NotFoundException')]. Final ListTables: D ['ackq_04fb03_t1',
  'ackq_04fb03_t2'], E ['ackq_04fb03_t1'].
  - ACK: custom_delete, synced.when, pre-delete-cleanup · ops: CreateTable, GetTableReplicationStatus,
    PutTableReplication, DeleteTableReplication, DeleteTableBucketReplication · fields:
    destinations[].destinationTableBucketArn, destinations[].replicationStatus
  - repro: bucket-level S->D + table-level t1->E -> CreateTable t2 -> status; Put t1 [D,E] -> status;
    DeleteTableReplication(t1) -> status x60 s; DeleteTableBucketReplication -> status x30 s
  - measurements: new_table_terminal_s=[2.13], n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-014](../tablebucket-replication.md#s3t-tablebucketreplication-014), [S3T-TABLEBUCKETREPLICATION-015](../tablebucket-replication.md#s3t-tablebucketreplication-015), [S3T-TABLEREPLICATION-017](../table-replication-status.md#s3t-tablereplication-017),
    [S3T-TABLEREPLICATION-042](../table-replication-status.md#s3t-tablereplication-042), [S3T-TABLEREPLICATION-021](../table-replication-status.md#s3t-tablereplication-021), [S3T-TABLEREPLICATION-025](../table-replication-status.md#s3t-tablereplication-025), [S3T-TABLEREPLICATION-020](../table-replication-status.md#s3t-tablereplication-020),
    [S3T-TABLEREPLICATION-027](../table-replication-status.md#s3t-tablereplication-027), [S3T-TABLEREPLICATION-043](../table-replication-status.md#s3t-tablereplication-043), [S3T-TABLE-138](../table-replication-status.md#s3t-table-138), [S3T-TABLEREPLICATION-018](../table-replication.md#s3t-tablereplication-018),
    [S3T-TABLEBUCKETREPLICATION-013](../tablebucket-replication.md#s3t-tablebucketreplication-013), [S3T-TABLEREPLICATION-047](../table-replication-status.md#s3t-tablereplication-047) · hypotheses: H-U-180 · evidence:
    table/sub-resources/repl-bucket-vs-table-config

## Notes

Interplay cells (n=1 each): (1) a table-level destination that is also the bucket-level destination does NOT
re-attach to the bucket-level replica - it fails with 'A table with an identical name already exists at the
destination.' while the other destination completes; (2) after DeleteTableReplication the per-table status
falls back to the bucket-level view ('D:failed' here - the failed entry persisted) after ~60 s of the stale
table-level view; (3) after DeleteTableBucketReplication both tables' status -> NotFoundException at once; (4)
a table created under the bucket-level config completed in 2.13 s and was not sent to another table's
table-level destination. Together with [S3T-TABLEREPLICATION-021](../table-replication-status.md#s3t-tablereplication-021): a table-level config REPLACES the
bucket-level destinations for that table (precedence, not union).

Contradiction with [S3T-TABLEREPLICATION-018](../table-replication.md#s3t-tablereplication-018): GetTableReplicationStatus after DeleteTableReplication: 018 ->
NotFoundException at +0.044 s with 0 stale 200 bodies; 022 -> HTTP 200 with the stale table-level view
('D:failed|E:completed') for ~60 s, then the bucket-level view ('D:failed'); the difference is state, not
direction: 022's table was also under bucket-level replication, so the status has a bucket view to fall back
to and only the table-level entries are stale Resolution: keep both; [S3T-TABLEREPLICATION-022](../table-replication-status.md#s3t-tablereplication-022) is the exception
for tables under bucket-level replication, [S3T-TABLEREPLICATION-018](../table-replication.md#s3t-tablereplication-018) holds without a bucket-level config -
018's title is scoped in title_fixes

Contradiction with [S3T-TABLEREPLICATION-047](../table-replication-status.md#s3t-tablereplication-047), [S3T-TABLEREPLICATION-017](../table-replication-status.md#s3t-tablereplication-017), [S3T-TABLEREPLICATION-042](../table-replication-status.md#s3t-tablereplication-042),
[S3T-TABLEREPLICATION-021](../table-replication-status.md#s3t-tablereplication-021): FU2-04 settled by follow-up probe
tablereplication/state-machine/fu2-04-put-offset-tick: 042/021/022 hold; 017's +1.06 s did not reproduce. The
delay is a fixed ~61 s per Put, not a periodic tick: Puts issued 0.9, 5, 20, 40, 55 and 70 s after the last
observed status change (67-565 s after the first Put) all showed the new destination as 'pending' 61.0-64.1 s
later and 'completed' 1.0-1.1 s after that (empty table, same-region buckets); the first Put behaved
identically (pendi Resolution: keep all; [S3T-TABLEREPLICATION-047](../table-replication-status.md#s3t-tablereplication-047) is canonical for the settled rule
