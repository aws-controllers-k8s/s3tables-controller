<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEREPLICATION-028: Two destinations, role allowed on B1 only (schema table): B2 'failed' at +66s, generic 'Insufficient permissions' message; B1 completed 475s
_Full entry and notes of one finding; its summary entry is in
[table-replication-status.md](../table-replication-status.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablereplication-028"></a>**S3T-TABLEREPLICATION-028** `async-state-machine` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Two destinations, role allowed on B1 only (schema table): B2 'failed' at +66s, generic 'Insufficient permissions' message; B1 completed 475s**
  Role: trust replication.s3tables.amazonaws.com, inline policy Allow s3tables:* on the source bucket and B1
  (bucket ARN and ARN/*), nothing on B2. PutTableReplication(destinations [B1, B2]) -> HTTP 200 {versionToken,
  status 'success'} on the first attempt (5 s after CreateRole). GetTableReplicationStatus every 5 s: +0 s
  destinations [] (no entries); +61 s B1 'pending' only (B2 entry not yet present); +66 s B1 'pending' + B2
  'failed'; +475 s B1 'completed' + B2 'failed'. Final B1 entry: {replicationStatus 'completed',
  destinationTableBucketArn, destinationTableArn, lastSuccessfulReplicatedUpdate {metadataLocation (source
  file), timestamp}}. Final B2 entry: {replicationStatus 'failed', destinationTableBucketArn, failureMessage
  'Insufficient permissions given to successfully complete replication.'} - no destinationTableArn, no
  lastSuccessfulReplicatedUpdate. ListTables: the replica exists in B1, none in B2. The B2 message is the same
  generic text seen for a nonexistent role / bad trust policy ([S3T-TABLEREPLICATION-002](../table-replication-status.md#s3t-tablereplication-002)/008): it names neither
  the denied action nor the destination bucket.
  - ACK: synced.when, requeue, hook:sdk_read_one_post_set_output · ops: PutTableReplication,
    GetTableReplicationStatus, ListTables · fields: destinations[].replicationStatus,
    destinations[].failureMessage, destinations[].destinationTableArn,
    destinations[].destinationTableBucketArn, configuration.role
  - repro: CreateRole(policy on source+B1) -> CreateTable(schema) -> PutTableReplication([B1,B2]) ->
    GetTableReplicationStatus every 5 s
  - measurements: b1_completed_after_put_s=[475.15], b2_failed_after_put_s=[65.72], put_attempts=[1], n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-021](../tablebucket-replication.md#s3t-tablebucketreplication-021), [S3T-TABLEBUCKET-011](../tablebucket-replication.md#s3t-tablebucket-011), [S3T-TABLEBUCKETREPLICATION-009](../tablebucket-replication.md#s3t-tablebucketreplication-009),
    [S3T-TABLEBUCKETREPLICATION-017](../tablebucket-replication.md#s3t-tablebucketreplication-017), [S3T-TABLEREPLICATION-032](../table-replication-status.md#s3t-tablereplication-032), [S3T-TABLEREPLICATION-002](../table-replication-status.md#s3t-tablereplication-002),
    [S3T-TABLEREPLICATION-008](../table-replication-status.md#s3t-tablereplication-008), [S3T-TABLEREPLICATION-033](../table-replication-status.md#s3t-tablereplication-033), [S3T-TABLEREPLICATION-005](../table-replication-status.md#s3t-tablereplication-005), [S3T-TABLEREPLICATION-019](../table-replication-status.md#s3t-tablereplication-019),
    [S3T-TABLEREPLICATION-017](../table-replication-status.md#s3t-tablereplication-017), [S3T-TABLEREPLICATION-029](../table-replication-status.md#s3t-tablereplication-029), [S3T-TABLEREPLICATION-031](../table-replication-status.md#s3t-tablereplication-031), [S3T-TABLEREPLICATION-006](../table-replication-status.md#s3t-tablereplication-006),
    [S3T-TABLEREPLICATION-043](../table-replication-status.md#s3t-tablereplication-043), [S3T-TABLEREPLICATION-027](../table-replication-status.md#s3t-tablereplication-027), [S3T-TABLEREPLICATION-020](../table-replication-status.md#s3t-tablereplication-020), [S3T-TABLEREPLICATION-023](../table-replication.md#s3t-tablereplication-023),
    [S3T-TABLEREPLICATION-024](../table-replication.md#s3t-tablereplication-024), [S3T-TABLEREPLICATION-013](../table-replication.md#s3t-tablereplication-013), [S3T-TABLEBUCKETREPLICATION-012](../tablebucket-replication.md#s3t-tablebucketreplication-012), [S3T-TABLE-026](../table-policy.md#s3t-table-026),
    [S3T-TABLEREPLICATION-041](../table-replication-status.md#s3t-tablereplication-041), [S3T-TABLEREPLICATION-025](../table-replication-status.md#s3t-tablereplication-025), [S3T-TABLEREPLICATION-030](../table-replication-status.md#s3t-tablereplication-030), [S3T-TABLE-136](../table-replication-status.md#s3t-table-136), [S3T-TABLE-224](../table-replication-status.md#s3t-table-224)
    · hypotheses: H-U-154 · evidence: tablereplication/state-machine/per-destination-failure

## Notes

H-U-154 confirmed for the per-destination part (independent statuses, the allowed destination completed while
the other stayed failed) and for the 30-min bound (failed at 66 s); REFUTED for 'message naming the denied
action' (generic text). Members are present-or-absent per entry (failed entries have no destinationTableArn).
n=1 table, schema-only (no data).

Full original title: Two destinations, role allowed on B1 only: per-destination status - B2 'failed' at +66 s
(generic message), B1 'pending' then 'completed' at +475 s
