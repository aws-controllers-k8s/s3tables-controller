<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEREPLICATION-008: Role trusting only ec2: PutTableReplication -> 200, 'failed' at ~61 s; no self-heal in 150 s after fixing trust; re-Put -> completed
_Full entry and notes of one finding; its summary entry is in
[table-replication-status.md](../table-replication-status.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablereplication-008"></a>**S3T-TABLEREPLICATION-008** `prerequisite` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Role trusting only ec2: PutTableReplication -> 200, 'failed' at ~61 s; no self-heal in 150 s after fixing trust; re-Put -> completed**
  With the replication role's trust policy changed to allow only ec2.amazonaws.com (UpdateAssumeRolePolicy ->
  HTTP 200, 30 s wait): PutTableReplication(new empty table -> D) -> HTTP 200; GetTableReplication -> HTTP 200.
  PutTableBucketReplication(B2, 10 tables -> D) with the same role -> HTTP 200. GetTableReplicationStatus(b8)
  every 5 s: 0.03s -> no-destinations; 60.66s -> failed:Insufficient permissions given to successfully
  complete replication.. Bucket-level: status(c1) -> no-destinations; new tables in D within the window: [].
  After restoring the trust policy (no new Put): status stayed 'failed' for 150 s; a re-Put of the identical
  configuration (current token) -> HTTP 200 and the status then completed after 65.88s.
  - ACK: requeue, synced.when, docs-only · ops: PutTableReplication, PutTableBucketReplication,
    GetTableReplicationStatus, UpdateAssumeRolePolicy · fields: configuration.role,
    destinations[].replicationStatus, destinations[].failureMessage
  - repro: UpdateAssumeRolePolicy(role, ec2 only) -> PutTableReplication / PutTableBucketReplication -> poll
    status -> restore trust -> poll -> re-Put -> poll
  - measurements: bad_trust_to_terminal_status_s=[60.66], self_heal_s=[], after_re_put_s=[65.88], n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-021](../tablebucket-replication.md#s3t-tablebucketreplication-021), [S3T-TABLEBUCKET-011](../tablebucket-replication.md#s3t-tablebucket-011), [S3T-TABLEBUCKETREPLICATION-009](../tablebucket-replication.md#s3t-tablebucketreplication-009),
    [S3T-TABLEBUCKETREPLICATION-017](../tablebucket-replication.md#s3t-tablebucketreplication-017), [S3T-TABLEREPLICATION-032](../table-replication-status.md#s3t-tablereplication-032), [S3T-TABLEREPLICATION-002](../table-replication-status.md#s3t-tablereplication-002),
    [S3T-TABLEREPLICATION-028](../table-replication-status.md#s3t-tablereplication-028), [S3T-TABLEBUCKETREPLICATION-012](../tablebucket-replication.md#s3t-tablebucketreplication-012), [S3T-TABLEBUCKETREPLICATION-028](../tablebucket-replication.md#s3t-tablebucketreplication-028),
    [S3T-TABLEREPLICATION-009](../table-replication.md#s3t-tablereplication-009), [S3T-TABLEBUCKETREPLICATION-027](../tablebucket-replication.md#s3t-tablebucketreplication-027), [S3T-TABLEBUCKETREPLICATION-013](../tablebucket-replication.md#s3t-tablebucketreplication-013),
    [S3T-TABLEBUCKETREPLICATION-030](../tablebucket-replication.md#s3t-tablebucketreplication-030), [S3T-TABLEBUCKETREPLICATION-002](../tablebucket-replication.md#s3t-tablebucketreplication-002), [S3T-TABLEREPLICATION-043](../table-replication-status.md#s3t-tablereplication-043),
    [S3T-TABLEREPLICATION-027](../table-replication-status.md#s3t-tablereplication-027), [S3T-TABLEREPLICATION-029](../table-replication-status.md#s3t-tablereplication-029), [S3T-TABLEREPLICATION-020](../table-replication-status.md#s3t-tablereplication-020) · hypotheses: H-U-131,
    H-U-132, H-S-053 · evidence: tablebucket/sub-resources/repl-backfill-delete

## Notes

H-U-131 confirmed: Put 200 (configuration stored), failure visible only as status 'failed' at ~61 s with the
generic message 'Insufficient permissions given to successfully complete replication.' (the message names
neither the role nor AssumeRole; the same text appears for a nonexistent role). H-U-132 REFUTED (no
synchronous rejection). H-S-053 confirmed at bucket level (Put 200 with the bad-trust role; no replicas
appeared). Extra: after the trust was fixed the status did NOT self-heal within 150 s, but a re-Put of the
identical configuration (current token) re-triggered replication and reached 'completed' 66 s later - the
controller fix for a failed prerequisite is a token-refreshing re-Put, not waiting.

Full original title: Role trusting only ec2: PutTableReplication -> 200; status 'failed' after 60.66s; after
fixing the trust the status did not self-heal in 150 s; after a re-Put: completed in 65.88s

Contradiction with [S3T-TABLEBUCKETREPLICATION-027](../tablebucket-replication.md#s3t-tablebucketreplication-027): 027's (truncated) title says an identical re-Put 'does NOT
re-trigger replication' - observed only for 'completed' tables (unchanged 150 s) and an in-flight backfill
(not restarted) under a bucket-level rule; TR-008 observed an identical table-level re-Put with the current
token re-trigger a 'failed' configuration to 'completed' in 65.88 s after the trust policy was fixed (waiting
150 s did not). 030 adds that a CHANGED document re-pends a 'completed' table. The direction depends on the
per-destination state (completed/in-flight untouched, failed re-attempted) and on whether the document
changed. Resolution: keep both; state-dependent, neither record is wrong - 027 over-generalizes in its
truncated title, see title_fixes
