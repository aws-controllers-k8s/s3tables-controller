<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKETREPLICATION-033: Reverify S3T-TABLEBUCKETREPLICATION-032: C2-C4 reproduce (identical-name 'failed' at +0.2/5.4 s) but the deleted replica IS re-created...
_Full entry and notes of one finding; its summary entry is in
[tablebucket-replication.md](../tablebucket-replication.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucketreplication-033"></a>**S3T-TABLEBUCKETREPLICATION-033** `delete-semantics` · impact high · unhandled (not handled in controller) · verified 2026-10-15
  **Reverify [S3T-TABLEBUCKETREPLICATION-032](../tablebucket-replication.md#s3t-tablebucketreplication-032): C2-C4 reproduce (identical-name 'failed' at +0.2/5.4 s) but the deleted replica IS re-created...**
  Three empty source tables under a bucket-level rule were listed in D after 16.0 s (all 'completed' after
  16.0 s). C1 DeleteTable(replica of ta in D) -> HTTP 204; GetTableReplicationStatus(ta) right after ->
  'completed'; after a 123.7 s poll the replica was re-created: True. C2 DeleteTable(source tb, tb2) -> HTTP
  204/204; status by the old ARNs -> ERR:NotFoundException/ERR:NotFoundException; replicas still in D at the
  end of the window: True/True. C3 CreateTable(same namespace/name as tb, tb2) -> HTTP 200/200 (new ARN
  differs: True/True); status at +0 s 'no-destinations'; first 'failed' at +0.2/5.4 s with message ['A table
  with an identical name already exists at the destination.']; the replica's ARN unchanged: True, its
  replicationInformation.sourceTableARN still the OLD source: True (points at the new source: False). C4
  control tables tc, td -> listed in D at +5.4/10.6 s, 'completed' at +5.4/10.6 s.
  - ACK: pre-delete-cleanup, requeue, e2e-timing · ops: PutTableBucketReplication, DeleteTable, CreateTable,
    GetTableReplicationStatus, ListTables, GetTable · fields: destinations.replicationStatus,
    destinations.failureMessage, managedTableInformation.replicationInformation.sourceTableARN
  - repro: PutTableBucketReplication(S->D) -> CreateTable ta,tb,tb2 (empty) -> wait listed in D ->
    DeleteTable(replica ta in D) + DeleteTable(source tb,tb2) -> poll D + statuses 120 s -> CreateTable(tb,tb2
    same names) + CreateTable(tc,td) -> poll 150 s -> GetTable(replica tb)
  - measurements: initial_listed_all_s=16.0, initial_completed_all_s=16.0, c1_c2_window_s=123.7,
    c3_c4_window_s=151.5, same_name_failed_at_s=[0.2, 5.4], control_listed_in_d_s=[5.4, 10.6],
    control_completed_s=[5.4, 10.6], n=2
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-032](../tablebucket-replication.md#s3t-tablebucketreplication-032) · evidence: tablebucketreplication/creative/reverify-set-d

## Notes

Reverify set D for [S3T-TABLEBUCKETREPLICATION-032](../tablebucket-replication.md#s3t-tablebucketreplication-032) with n=2 for the same-name failure and the control timing
(original n=1); poll interval 5 s so +X s values are upper bounds within 5 s. CORRECTS clause C1 ('deleted
replica not re-created within 120 s'): the replica of ta reappeared in ListTables(D) at +46.7 s while the
source's status stayed 'completed' at every 5 s sample; two other sources were deleted in the same second, so
the follow-up probe tablebucketreplication/creative/reverify-set-d-c1 isolates the trigger (re-created at
+31.1/36.2 s with no source activity). Initial listing of three empty tables took 16.0 s here vs 61 s in the
original.

Full original title: Reverify [S3T-TABLEBUCKETREPLICATION-032](../tablebucket-replication.md#s3t-tablebucketreplication-032): C2-C4 reproduce (identical-name 'failed' at
+0.2/5.4 s) but the deleted replica IS re-created (+46.7 s)
