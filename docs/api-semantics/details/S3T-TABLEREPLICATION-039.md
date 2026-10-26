<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEREPLICATION-039: 10 identical PutTableReplication re-Puts back-to-back (fresh token each): 10 x 200 in 1.0 s, token rotates each time, status stays completed
_Full entry and notes of one finding; its summary entry is in [table-replication.md](../table-replication.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-tablereplication-039"></a>**S3T-TABLEREPLICATION-039** `quota-limit` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **10 identical PutTableReplication re-Puts back-to-back (fresh token each): 10 x 200 in 1.0 s, token rotates each time, status stays completed**
  On a 'completed' table-level configuration: 10 sequential identical re-Puts, each sending the token returned
  by the previous response: codes {"HTTP 200": 10}, latencies [102, 102, 106, 107, 95, 109, 99, 101, 97, 92]
  ms, wall 1.02s; token rotated on 10 of the successful calls; GetTableReplication afterwards returns the last
  token: True. GetTableReplicationStatus right after the burst: 'completed'; 12 samples over the next 60 s:
  ['completed']. Destination bucket still lists ['ackq_5c8005_t'] (no duplicate replica).
  - ACK: custom_update, annotation-shadow-state, requeue · ops: PutTableReplication, GetTableReplication,
    GetTableReplicationStatus · fields: versionToken, destinations[].replicationStatus
  - repro: completed table-level replication -> 10 x PutTableReplication(same configuration, token from the
    previous response) with no pause -> status samples
  - measurements: storm_wall_s=1.02, latency_ms=[102, 102, 106, 107, 95, 109, 99, 101, 97, 92], n=10
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEREPLICATION-038](../table-replication.md#s3t-tablereplication-038), [S3T-TABLEREPLICATION-016](../table-replication.md#s3t-tablereplication-016), [S3T-TABLEBUCKETREPLICATION-002](../tablebucket-replication.md#s3t-tablebucketreplication-002),
    [S3T-TABLEREPLICATION-044](../table-replication.md#s3t-tablereplication-044), [S3T-TABLEREPLICATION-014](../table-replication.md#s3t-tablereplication-014), [S3T-TABLEREPLICATION-029](../table-replication-status.md#s3t-tablereplication-029), [S3T-TABLEREPLICATION-003](../table-replication.md#s3t-tablereplication-003) ·
    evidence: tablereplication/creative/reput-storm-and-lost-destination

## Notes

Cross-service analogy (DynamoDB SSE/TTL change cooldowns, DeletionProtection 15 s cooldown; ElastiCache no-op
rejection): is a controller that re-Puts the replication configuration every reconcile (because the token
rotates, [S3T-TABLEREPLICATION-038](../table-replication.md#s3t-tablereplication-038)) throttled, rejected or does it reset the replication? Extends
[S3T-TABLEREPLICATION-015](../table-replication.md#s3t-tablereplication-015)/016 (stale/concurrent tokens) with the sequential fresh-token case. n=10 calls, one
table.
