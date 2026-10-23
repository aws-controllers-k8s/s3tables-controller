<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKETREPLICATION-032: Bucket rule: source delete leaves the replica; same-name re-create -> failed 'identical name' +5 s (replica re-creation: -033/-034)
_Full entry and notes of one finding; its summary entry is in
[tablebucket-replication.md](../tablebucket-replication.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucketreplication-032"></a>**S3T-TABLEBUCKETREPLICATION-032** `delete-semantics` · impact high · unhandled (not handled in controller) · verified 2026-10-15, re-verified
  **Bucket rule: source delete leaves the replica; same-name re-create -> failed 'identical name' +5 s (replica re-creation: -033/-034)**
  Two empty source tables ta, tb replicated to D under PutTableBucketReplication (D listed both after 61.0s).
  C1 DeleteTable(replica of ta in D) -> HTTP 204; GetTableReplicationStatus(ta) right after 'completed'; 120 s
  poll (D tables, status): [[0.0, "completed", ["ackq_97f768_tb@ff92f100"]]]. C2 DeleteTable(source tb) ->
  HTTP 204; status by tb's ARN right after: 'ERR:NotFoundException'; 120 s poll: [[0.0,
  "ERR:NotFoundException", ["ackq_97f768_tb@ff92f100"]]]. C3 CreateTable(same namespace/name as tb) -> HTTP
  200 (new ARN: True); 150 s poll of the new table's status + D: [[0.1, "no-destinations",
  ["ackq_97f768_tb@ff92f100"]], [5.2, "failed:A table with an identical name already exists at the
  destination.", ["ackq_97f768_tb@ff92f100"]]]; replica entry before/after {"before":
  ["ackq_97f768_tb@ff92f100"], "after": ["ackq_97f768_tb@ff92f100"]}; the replica's
  replicationInformation.sourceTableARN now points to the old source. C4 control CreateTable tc -> listed in D
  after 5.2s, status 'completed'.
  - ACK: pre-delete-cleanup, requeue, custom_delete · ops: PutTableBucketReplication, DeleteTable,
    CreateTable, GetTableReplicationStatus, ListTables, GetTable · fields: destinations[].replicationStatus,
    destinations[].failureMessage, managedTableInformation.replicationInformation.sourceTableARN
  - repro: PutTableBucketReplication -> 2 tables replicated -> DeleteTable(replica) / DeleteTable(source) /
    CreateTable(same name) / CreateTable(new) with 5 s polling of D and status
  - measurements: c1_window_s=121.7, c2_window_s=121.4, c3_window_s=152.1, c4_listed_s=5.2, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-138](../table-replication-status.md#s3t-table-138), [S3T-TABLE-196](../table-replication-status.md#s3t-table-196), [S3T-TABLEREPLICATION-020](../table-replication-status.md#s3t-tablereplication-020), [S3T-TABLEBUCKETREPLICATION-014](../tablebucket-replication.md#s3t-tablebucketreplication-014),
    [S3T-TABLEBUCKETREPLICATION-027](../tablebucket-replication.md#s3t-tablebucketreplication-027), [S3T-TABLEBUCKETREPLICATION-028](../tablebucket-replication.md#s3t-tablebucketreplication-028), [S3T-TABLEBUCKETREPLICATION-031](../tablebucket-replication.md#s3t-tablebucketreplication-031),
    [S3T-TABLEBUCKETREPLICATION-015](../tablebucket-replication.md#s3t-tablebucketreplication-015), [S3T-TABLEBUCKETREPLICATION-016](../tablebucket-replication.md#s3t-tablebucketreplication-016), [S3T-TABLEREPLICATION-007](../table-replication.md#s3t-tablereplication-007),
    [S3T-TABLEBUCKETREPLICATION-005](../tablebucket-replication.md#s3t-tablebucketreplication-005), [S3T-TABLE-224](../table-replication-status.md#s3t-table-224) · evidence:
    tablebucketreplication/creative/replica-delete-and-source-recreate

## Notes

Cross-service analogy (DynamoDB: deleting a replica region is explicit and the table name is shared; a
delete+recreate of a CR with the same name is the classic ACK adoption loop). Extends [S3T-TABLE-138](../table-replication-status.md#s3t-table-138)
(table-level: deleted replica not re-created; later commit -> failed 'destination table has been deleted'),
[S3T-TABLE-196](../table-replication-status.md#s3t-table-196) (table-level: replicas outlive the source) and [S3T-TABLEREPLICATION-020](../table-replication-status.md#s3t-tablereplication-020)/025 (name collision ->
failed) to the bucket-level rule. Empty tables, same-region destination, no commits, n=1 per cell. Controller
trap: an ACK Table deleted and re-created with the same spec (name/namespace) under a bucket-level replication
rule gets a NEW ARN and a permanently 'failed' replication status ('A table with an identical name already
exists at the destination.') because the orphaned replica still points at the OLD source ARN; the status by
the old ARN is NotFoundException. Recovery needs an out-of-band DeleteTable of the stale replica in the
destination bucket (which, per C1, the rule does not re-create by itself).

Re-verification CORRECTED (2026-10-15, independent probe tablebucketreplication/creative/reverify-set-d,
tablebucketreplication/creative/reverify-set-d-c1): C2-C4 reproduce: source DeleteTable -> 204 with status by
old ARN NotFoundException and the replica left in D (2/2, 123 s); same-name re-create -> 200 new ARN, status
'no-destinations' then 'failed' 'A table with an identical name already exists at the destination.' at
+0.2/+5.4 s (n=2), replica ARN unchanged and sourceTableARN still the old source; controls listed in D and
'completed' at +5.4/+10.6 s (n=2). C1 differs: the replica deleted in D IS re-created - at +46.7 s in the main
probe and at +31.1/+36.2 s with no source activity in the isolating probe (status completed -> pending ->
completed), so 'deleted replica not re-created (120 s)' does not hold (n=3). Recorded as
[S3T-TABLEBUCKETREPLICATION-033](../tablebucket-replication.md#s3t-tablebucketreplication-033)/-034.
