<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEREPLICATION-043: Corrective re-Put on a 'failed' collision dest: stale 'failed'+old message 61 s, then 'completed' directly; no self-heal without Put (300 s)
_Full entry and notes of one finding; its summary entry is in
[table-replication-status.md](../table-replication-status.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablereplication-043"></a>**S3T-TABLEREPLICATION-043** `stale-response` · impact high · unhandled (not handled in controller) · verified 2026-10-14
  **Corrective re-Put on a 'failed' collision dest: stale 'failed'+old message 61 s, then 'completed' directly; no self-heal without Put (300 s)**
  Destination collision (customer table with the replica's name): PutTableReplication -> HTTP 200 status
  'success'; GetTableReplicationStatus at +59 ms -> []; then []@0.03s -> failed+msg@61.15s. Deleting the
  colliding table alone: status stayed 'failed' with the old message for 320 s (a 300 s watch in an earlier
  run of this probe plus 20 s here) - no self-heal. Re-Put of the unchanged configuration with the current
  token -> HTTP 200 (token rotated: True); status right after (+0.06 s) -> failed+msg; 1 s polling:
  failed+msg@0.02s -> completed@61.17s. A controller that reads the status right after its re-Put sees the
  stale 'failed' with the old failureMessage and would wrongly report failure (or re-Put again).
  - ACK: requeue, synced.when, custom_update · ops: PutTableReplication, GetTableReplicationStatus,
    DeleteTable · fields: destinations.replicationStatus, destinations.failureMessage, versionToken
  - repro: Create table in D with the source's ns/name -> PutTableReplication -> watch failed ->
    DeleteTable(D) -> wait 20 s -> PutTableReplication(current token) -> poll status 1/s
  - measurements: failed_to_first_at_s=61.15, stale_failed_after_reput_s=61.17, completed_after_reput_s=61.17,
    n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEREPLICATION-025](../table-replication-status.md#s3t-tablereplication-025), [S3T-TABLEREPLICATION-026](../table-replication-status.md#s3t-tablereplication-026), [S3T-TABLEREPLICATION-027](../table-replication-status.md#s3t-tablereplication-027),
    [S3T-TABLEREPLICATION-019](../table-replication-status.md#s3t-tablereplication-019), [S3T-TABLEREPLICATION-031](../table-replication-status.md#s3t-tablereplication-031), [S3T-TABLEREPLICATION-038](../table-replication.md#s3t-tablereplication-038), [S3T-TABLEREPLICATION-030](../table-replication-status.md#s3t-tablereplication-030),
    [S3T-TABLEREPLICATION-005](../table-replication-status.md#s3t-tablereplication-005), [S3T-TABLEBUCKETREPLICATION-013](../tablebucket-replication.md#s3t-tablebucketreplication-013), [S3T-TABLEREPLICATION-020](../table-replication-status.md#s3t-tablereplication-020),
    [S3T-TABLEREPLICATION-022](../table-replication-status.md#s3t-tablereplication-022), [S3T-TABLE-138](../table-replication-status.md#s3t-table-138), [S3T-TABLEREPLICATION-029](../table-replication-status.md#s3t-tablereplication-029), [S3T-TABLEREPLICATION-028](../table-replication-status.md#s3t-tablereplication-028),
    [S3T-TABLEREPLICATION-008](../table-replication-status.md#s3t-tablereplication-008) · evidence: tablereplication/creative/controller-loop

## Notes

extends [S3T-TABLEREPLICATION-026](../table-replication-status.md#s3t-tablereplication-026)/-027 with what the status shows BETWEEN the corrective Put and 'completed'.
Decision rule: after any Put, ignore 'failed' entries for at least the measured stale window (or until the
entry changes) before setting a failure condition; 'completed' is the only state that proves the new Put was
evaluated. Single run, same-region destination.

Full original title: Corrective re-Put on a 'failed' destination: status shows the stale 'failed'+old message
for 61 s, then 'completed' directly; no self-heal without a Put (300 s)
