<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-306: RenameTable (name, then namespace) on a replicated source: replica keeps the OLD name/namespace for 490 s; status completed; re-Put no-op
_Full entry and notes of one finding; its summary entry is in
[table-replication-status.md](../table-replication-status.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-306"></a>**S3T-TABLE-306** `requested-vs-effective` · impact high · unhandled (not handled in controller) · verified 2026-10-14
  **RenameTable (name, then namespace) on a replicated source: replica keeps the OLD name/namespace for 490 s; status completed; re-Put no-op**
  Empty source table t replicated to D (completed at +66s; D view {"ackq_7b3249_ns": ["ackq_7b3249_t"]}).
  RenameTable(newName=t2) -> HTTP 204. Over 120 s polling D every 5 s: transitions [[0.1, "completed",
  {"ackq_7b3249_ns": ["ackq_7b3249_t"]}]]; final D view {"ackq_7b3249_ns": ["ackq_7b3249_t"]};
  GetTable(replica ARN).name == new name: False. Then identical PutTableReplication (current token) -> HTTP
  200, status right after 'completed'; over 245s: [[0.1, "completed", {"ackq_7b3249_ns": ["ackq_7b3249_t"]}]];
  final D view {"ackq_7b3249_ns": ["ackq_7b3249_t"]}; final status entries [{"replicationStatus": "completed",
  "destinationTableArn":
  "arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-7b3249-rfr-dst/table/b3c876c2-2d35-4a11-b88c-a185d682f40e"}].
  Then RenameTable(newNamespaceName=ns2) -> HTTP 204; over 122s: [[0.1, "completed", {"ackq_7b3249_ns":
  ["ackq_7b3249_t"]}]]; final D view {"ackq_7b3249_ns": ["ackq_7b3249_t"]}.
  - ACK: custom_update, pre-delete-cleanup, requeue · ops: RenameTable, PutTableReplication,
    GetTableReplicationStatus, ListTables, GetTable · fields: newName, newNamespaceName,
    destinations[].destinationTableArn, destinations[].replicationStatus
  - repro: PutTableReplication(empty table -> D) -> completed -> RenameTable(newName) -> poll D 120 s ->
    re-Put identical -> poll 240 s -> RenameTable(newNamespaceName) -> poll 120 s
  - measurements: initial_completed_s=65.79, r1_window_s=122.3, r2_window_s=244.6, r3_window_s=122.3, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-224](../table-replication-status.md#s3t-table-224), [S3T-TABLEREPLICATION-020](../table-replication-status.md#s3t-tablereplication-020), [S3T-TABLE-136](../table-replication-status.md#s3t-table-136), [S3T-TABLE-299](../table-replication-status.md#s3t-table-299), [S3T-TABLE-300](../table-replication-status.md#s3t-table-300),
    [S3T-TABLE-138](../table-replication-status.md#s3t-table-138), [S3T-TABLEBUCKET-118](../tablebucket-replication.md#s3t-tablebucket-118) · evidence: table/creative/replica-follows-rename

## Notes

Cross-service analogy (DynamoDB replicas are addressed by the same table name; SNS has no rename): the
controller's only Update (RenameTable) interacts with a service-managed replica. Extends [S3T-TABLE-224](../table-replication-status.md#s3t-table-224) (rename
of the replica itself is allowed) and [S3T-TABLEREPLICATION-020](../table-replication-status.md#s3t-tablereplication-020) (re-Put with a leftover replica -> failed).
Empty table, same-region destination, config-only (no commit) during the windows; a commit-driven sync (~7
min, [S3T-TABLEREPLICATION-006](../table-replication-status.md#s3t-tablereplication-006)) was not waited for. n=1. The 'same namespace/name as the source' relation of
[S3T-TABLE-136](../table-replication-status.md#s3t-table-136) is therefore a creation-time property: after a rename the replica can only be located through
GetTableReplicationStatus.destinations[].destinationTableArn (which kept pointing at the old-named replica).
