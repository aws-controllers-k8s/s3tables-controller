<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKETREPLICATION-034: Bucket rule: a replica deleted in D IS re-created with no source activity, at +31.1/36.2 s (n=2, 5 s polls)
_Full entry and notes of one finding; its summary entry is in
[tablebucket-replication.md](../tablebucket-replication.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucketreplication-034"></a>**S3T-TABLEBUCKETREPLICATION-034** `delete-semantics` · impact high · unhandled (not handled in controller) · verified 2026-10-15
  **Bucket rule: a replica deleted in D IS re-created with no source activity, at +31.1/36.2 s (n=2, 5 s polls)**
  Two empty sources ta1, ta2 replicated to D (listed+completed after 10.7 s), then 20 s quiet.
  DeleteTable(replica ta1/ta2 in D) -> HTTP 204/204; GetTableReplicationStatus(source) right after ->
  'completed'/'completed'. Replica reappearance in ListTables(D): quiet 180 s: reappeared {'ta1': 31.1, 'ta2':
  36.2}. Status strings seen during the quiet window: ['completed', 'pending'].
  - ACK: pre-delete-cleanup, requeue · ops: DeleteTable, ListTables, GetTableReplicationStatus,
    PutTableBucketReplication, CreateTable · fields: destinations.replicationStatus
  - repro: PutTableBucketReplication(S->D) -> CreateTable ta1,ta2 -> wait completed -> DeleteTable(replicas in
    D) -> poll D 180 s quiet -> (if absent) CreateTable tx -> poll 120 s -> DeleteTable tx -> poll 120 s ->
    re-Put config -> poll 120 s
  - measurements: quiet_reappeared_at_s.ta1=31.1, quiet_reappeared_at_s.ta2=36.2,
    reappeared_by_phase.ta1.phase=p1, reappeared_by_phase.ta1.at_s=31.1, reappeared_by_phase.ta2.phase=p1,
    reappeared_by_phase.ta2.at_s=36.2, n=2, poll_interval_s=5
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-032](../tablebucket-replication.md#s3t-tablebucketreplication-032) · evidence: tablebucketreplication/creative/reverify-set-d-c1

## Notes

Reverify set D follow-up isolating clause C1 of [S3T-TABLEBUCKETREPLICATION-032](../tablebucket-replication.md#s3t-tablebucketreplication-032) (in the main reverify probe the
replica reappeared at +46.7 s while two other sources were deleted in the same second). REFUTES clause C1 of
[S3T-TABLEBUCKETREPLICATION-032](../tablebucket-replication.md#s3t-tablebucketreplication-032): with n=3 across the two probes the deleted replica reappeared at +31.1, +36.2
and +46.7 s; the source status flipped completed -> pending -> completed around the re-creation (sampled at
31.1 s), i.e. the service re-replicates the table rather than leaving the destination deletion in place.
