<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEREPLICATION-045: GetTableReplicationStatus right after PutTableReplication -> 200 with destinations [] for ~60 s...
_Full entry and notes of one finding; its summary entry is in
[table-replication-status.md](../table-replication-status.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablereplication-045"></a>**S3T-TABLEREPLICATION-045** `eventual-consistency` · impact medium · unhandled (not handled in controller) · verified 2026-10-15
  **GetTableReplicationStatus right after PutTableReplication -> 200 with destinations [] for ~60 s...**
  PutTableReplication -> HTTP 200 status 'success'; GetTableReplicationStatus at +0 s -> HTTP 200
  {sourceTableArn, destinations: []} (no destination entry and no error). Polled every 5 s, the destination
  entry appeared with replicationStatus 'pending' (destinationTableBucketArn and destinationTableArn present,
  no lastSuccessfulReplicatedUpdate) at 60.7 s and was still 'pending' at 300 s, while ListTables on the
  destination bucket already listed the replica (same namespace and table name as the source, type
  'customer').
  - ACK: requeue, hook:sdk_read_one_post_set_output · ops: PutTableReplication, GetTableReplicationStatus,
    ListTables · fields: destinations, destinations.replicationStatus, destinations.destinationTableArn
  - repro: CreateTable(schema) -> PutTableReplication -> GetTableReplicationStatus immediately, then every 5 s
    for 300 s; ListTables(destination)
  - measurements: status_destinations_empty_for_s=60.7, still_pending_at_s=300, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEREPLICATION-003](../table-replication.md#s3t-tablereplication-003) · evidence: tablereplication/creative/reverify-set-e-repl

## Notes

Precision gain while re-verifying [S3T-TABLEREPLICATION-003](../table-replication.md#s3t-tablereplication-003) (whose Put/Get clauses all reproduced): a
controller reading status right after Put sees an empty destinations list, not pending; single run, timings
are one sample.

Full original title: GetTableReplicationStatus right after PutTableReplication -> 200 with destinations [] for
~60 s; destination then 'pending' for >= 300 s although the replica table already exists
