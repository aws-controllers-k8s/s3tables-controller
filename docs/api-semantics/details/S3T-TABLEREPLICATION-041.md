<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEREPLICATION-041: Replicated source: RenameTable/namespace move leave replica at old ns/name 150 s, status 'completed'; tags/policy/maintenance absent at +5s
_Full entry and notes of one finding; its summary entry is in
[table-replication-status.md](../table-replication-status.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablereplication-041"></a>**S3T-TABLEREPLICATION-041** `async-state-machine` · impact high · unhandled (not handled in controller) · verified 2026-10-14
  **Replicated source: RenameTable/namespace move leave replica at old ns/name 150 s, status 'completed'; tags/policy/maintenance absent at +5s**
  Replication completed after 66.1 s (same-region destination). A)
  TagResource/PutTablePolicy/PutTableMaintenanceConfiguration on the SOURCE -> HTTP 200 / HTTP 200 / HTTP 204;
  5 s later the replica reads tags {"tags": {}}, policy NotFoundException/404 'The specified table policy does
  not exist.', snapshot maintenance {"status": "enabled", "settings": {"icebergSnapshotManagement":
  {"minSnapshotsToKeep": 1, "maxSnapshotAgeHours": 120}}} (status completed). B) RenameTable(source, newName)
  -> HTTP 204; destination/status over 150 s (samples deduplicated; only one distinct state seen): +0.1s:
  D={"ackq_5ad769_ns": ["ackq_5ad769_t"], "ackq_5ad769_ns2": "ERR:NotFoundException"} status=completed
  replica={"managedByService": "replication.s3tables.amazonaws.com", "name": "ackq_5ad769_t", "namespace":
  ["ackq_5ad769_ns"]}. C) RenameTable(source, newNamespaceName=ns2) with ns2 existing only in the source
  bucket -> HTTP 204; over 150 s (one distinct state): +0.1s: D={"ackq_5ad769_ns": ["ackq_5ad769_t"],
  "ackq_5ad769_ns2": "ERR:NotFoundException"} status=completed replica={"managedByService":
  "replication.s3tables.amazonaws.com", "name": "ackq_5ad769_t", "namespace": ["ackq_5ad769_ns"]}; namespaces
  in D at the end: ['ackq_5ad769_ns']. In both cases the replica table kept its old namespace/name and
  managedByService, no new replica or namespace appeared in D and the status never left 'completed' (no
  metadata commit was made on the source).
  - ACK: custom_update, requeue, synced.when, is_immutable · ops: RenameTable, TagResource, PutTablePolicy,
    PutTableMaintenanceConfiguration, GetTableReplicationStatus, ListTables · fields: name, namespace, tags,
    resourcePolicy, icebergSnapshotManagement
  - repro: replicated empty table -> tag/policy/maintenance on source -> read on replica; RenameTable(newName)
    -> watch D/status 150 s; RenameTable(newNamespaceName=ns2 absent in D) -> watch 150 s
  - measurements: completed_at_s=66.1, rename_watch_last_at_s=0.1, move_watch_last_at_s=0.1, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-224](../table-replication-status.md#s3t-table-224), [S3T-TABLE-136](../table-replication-status.md#s3t-table-136), [S3T-TABLEREPLICATION-025](../table-replication-status.md#s3t-tablereplication-025), [S3T-TABLEREPLICATION-030](../table-replication-status.md#s3t-tablereplication-030),
    [S3T-TABLEREPLICATION-031](../table-replication-status.md#s3t-tablereplication-031), [S3T-TABLEREPLICATION-019](../table-replication-status.md#s3t-tablereplication-019), [S3T-TABLEREPLICATION-028](../table-replication-status.md#s3t-tablereplication-028) · evidence:
    tablereplication/creative/source-updates-vs-replica

## Notes

extends [S3T-TABLE-224](../table-replication-status.md#s3t-table-224)/-136 (what a replica accepts) with the SOURCE-side view a Table controller has: none of
its ordinary updates (tags, policy, maintenance, rename, namespace move) reach the replica, and the
replication status keeps reporting 'completed' although destination and source no longer agree on
namespace/name. Decision rule: treat the replica's identity and sub-resources as independent resources; a
spec.name/namespace change on a replicated source cannot be considered propagated by watching
GetTableReplicationStatus; whether the next metadata commit creates a second replica under the new name was
not tested (no commit issued). Single run, 150 s windows, same-region destination.

Full original title: Replicated source: RenameTable and a namespace move leave the replica at the old ns/name
for 150 s with status 'completed'; tags/policy/maintenance never propagate
