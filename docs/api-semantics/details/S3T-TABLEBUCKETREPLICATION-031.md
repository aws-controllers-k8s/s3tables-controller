<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKETREPLICATION-031: Bucket-level replication ignores RenameTable (name, namespace) on a source for 150/120 s: replica keeps the old name; new table in 5 s
_Full entry and notes of one finding; its summary entry is in
[tablebucket-replication.md](../tablebucket-replication.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucketreplication-031"></a>**S3T-TABLEBUCKETREPLICATION-031** `requested-vs-effective` · impact high · unhandled (not handled in controller) · verified 2026-10-14
  **Bucket-level replication ignores RenameTable (name, namespace) on a source for 150/120 s: replica keeps the old name; new table in 5 s**
  PutTableBucketReplication(S -> D) -> HTTP 200; then CreateNamespace ns + CreateTable t in S: replica listed
  in D after 66.0 (status completed). RenameTable(t -> t2) -> HTTP 204; D polled every 5 s for 153s:
  transitions [[0.1, "completed", {"ackq_6c3e6e_ns": ["ackq_6c3e6e_t"]}]]; final D {"ackq_6c3e6e_ns":
  ["ackq_6c3e6e_t"]}; GetTable(replica ARN) {"name": "ackq_6c3e6e_t", "namespace": ["ackq_6c3e6e_ns"]};
  GetTableReplicationStatus(source ARN) -> 'completed'. Control CreateTable t3 -> listed in D after 5.2s
  (status transitions [[0.1, "no-destinations"], [5.2, "completed"]]). RenameTable(t2 -> ns2/t2) -> HTTP 204;
  after 122s D = {"ackq_6c3e6e_ns": ["ackq_6c3e6e_t", "ackq_6c3e6e_t3"]} (transitions [[0.1, "completed",
  {"ackq_6c3e6e_ns": ["ackq_6c3e6e_t", "ackq_6c3e6e_t3"]}]]).
  - ACK: custom_update, pre-delete-cleanup, requeue · ops: PutTableBucketReplication, RenameTable,
    CreateTable, ListTables, GetTableReplicationStatus · fields: newName, newNamespaceName,
    destinations[].replicationStatus
  - repro: PutTableBucketReplication -> CreateTable -> replica appears -> RenameTable(newName) -> poll D 150 s
    -> CreateTable (control) -> RenameTable(newNamespaceName) -> poll D 120 s
  - measurements: initial_replica_listed_s=66.0, b1_window_s=153.0, control_new_table_listed_s=5.2,
    b3_window_s=122.3, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-014](../tablebucket-replication.md#s3t-tablebucketreplication-014), [S3T-TABLE-224](../table-replication-status.md#s3t-table-224), [S3T-TABLEREPLICATION-020](../table-replication-status.md#s3t-tablereplication-020),
    [S3T-TABLEBUCKETREPLICATION-027](../tablebucket-replication.md#s3t-tablebucketreplication-027), [S3T-TABLEBUCKETREPLICATION-028](../tablebucket-replication.md#s3t-tablebucketreplication-028), [S3T-TABLEBUCKETREPLICATION-032](../tablebucket-replication.md#s3t-tablebucketreplication-032) · evidence:
    tablebucketreplication/creative/rename-under-bucket-replication

## Notes

Cross-service analogy (DynamoDB replicas are addressed by the shared table name; bucket-level replication
replicates new tables within seconds per [S3T-TABLEBUCKETREPLICATION-014](../tablebucket-replication.md#s3t-tablebucketreplication-014)): the controller's only Update
(RenameTable) against a bucket-wide replication rule. Compare the table-level case in
table/creative/replica-follows-rename. Empty tables, same-region destination, no commits, n=1. Timing detail:
the table created ~1 s after PutTableBucketReplication was listed in D only after 66 s (status
'no-destinations' -> pending at 60.9 s -> completed at 66.0 s), while the control table created ~4 min later
was listed after 5.2 s - the ~60 s warm-up of [S3T-TABLEREPLICATION-019](../table-replication-status.md#s3t-tablereplication-019) belongs to the configuration, not to
each table (compare [S3T-TABLEBUCKETREPLICATION-014](../tablebucket-replication.md#s3t-tablebucketreplication-014), 4.13 s).
