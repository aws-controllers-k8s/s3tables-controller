<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEREPLICATION-002: Nonexistent role: PutTableReplication -> 200; status 'failed: Insufficient permissions given to successfully complete replication.' ~61 s
_Full entry and notes of one finding; its summary entry is in
[table-replication-status.md](../table-replication-status.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablereplication-002"></a>**S3T-TABLEREPLICATION-002** `prerequisite` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Nonexistent role: PutTableReplication -> 200; status 'failed: Insufficient permissions given to successfully complete replication.' ~61 s** (hypothesis refuted; behavior confirmed)
  PutTableReplication(configuration.role = ARN of a role that does not exist, existing destination bucket) ->
  HTTP 200. GetTableReplicationStatus polled every 5 s: 0.05s -> no-destinations; 60.61s ->
  failed:Insufficient permissions given to successfully complete replication.; final destinations[0] =
  'failed:Insufficient permissions given to successfully complete replication.'.
  - ACK: requeue, synced.when, references · ops: PutTableReplication, GetTableReplicationStatus · fields:
    configuration.role, destinations[].replicationStatus, destinations[].failureMessage
  - repro: PutTableReplication(role=arn of a missing role) -> poll GetTableReplicationStatus
  - measurements: nonexistent_role_to_terminal_status_s=[60.61], n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-021](../tablebucket-replication.md#s3t-tablebucketreplication-021), [S3T-TABLEBUCKET-011](../tablebucket-replication.md#s3t-tablebucket-011), [S3T-TABLEBUCKETREPLICATION-009](../tablebucket-replication.md#s3t-tablebucketreplication-009),
    [S3T-TABLEBUCKETREPLICATION-017](../tablebucket-replication.md#s3t-tablebucketreplication-017), [S3T-TABLEREPLICATION-032](../table-replication-status.md#s3t-tablereplication-032), [S3T-TABLEREPLICATION-008](../table-replication-status.md#s3t-tablereplication-008),
    [S3T-TABLEREPLICATION-028](../table-replication-status.md#s3t-tablereplication-028), [S3T-TABLEBUCKETREPLICATION-012](../tablebucket-replication.md#s3t-tablebucketreplication-012), [S3T-TABLEBUCKETREPLICATION-028](../tablebucket-replication.md#s3t-tablebucketreplication-028),
    [S3T-TABLEREPLICATION-009](../table-replication.md#s3t-tablereplication-009) · hypotheses: H-U-133, H-S-054 · evidence:
    service/error-taxonomy/repl-permissions

## Notes

H-U-133 second half (nonexistent role -> 400) REFUTED: accepted with 200; H-S-054 at table level likewise
refuted. Role validation is asynchronous: the status goes pending (destinations empty at +0) -> failed at
+60.6 s with the generic message 'Insufficient permissions given to successfully complete replication.' (the
message does not name the role or AssumeRole). n=1.

Full original title: PutTableReplication accepts a nonexistent role (200); status turns 'failed: Insufficient
permissions given to successfully complete replication.' at ~60 s
