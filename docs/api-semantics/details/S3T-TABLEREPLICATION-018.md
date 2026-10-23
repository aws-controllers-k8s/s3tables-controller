<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEREPLICATION-018: DeleteTableReplication(current token), no bucket-level config -> 204; Get/Status 404 at +0.019s/+0.044s (0 stale 200s); Delete again -> 404
_Full entry and notes of one finding; its summary entry is in [table-replication.md](../table-replication.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-tablereplication-018"></a>**S3T-TABLEREPLICATION-018** `delete-semantics` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **DeleteTableReplication(current token), no bucket-level config -> 204; Get/Status 404 at +0.019s/+0.044s (0 stale 200s); Delete again -> 404**
  DeleteTableReplication with the current token -> HTTP 204. Polling GetTableReplication and
  GetTableReplicationStatus alternately every ~200 ms: first NotFoundException at +0.019s (Get) and +0.044s
  (Status); stale HTTP 200 bodies seen: {'get': 0, 'status': 0}. A second Delete with the same (consumed)
  token -> NotFoundException (HTTP 404) 'The specified replication configuration does not exist.'.
  - ACK: custom_delete, exceptions.404 · ops: DeleteTableReplication, GetTableReplication,
    GetTableReplicationStatus · fields: versionToken
  - repro: Delete(current token) -> poll both Gets @200 ms -> Delete(same token)
  - measurements: get_404_after_delete_s=[0.019], status_404_after_delete_s=[0.044], n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEREPLICATION-007](../table-replication.md#s3t-tablereplication-007), [S3T-TABLEBUCKETREPLICATION-024](../tablebucket-replication.md#s3t-tablebucketreplication-024), [S3T-TABLEBUCKETREPLICATION-025](../tablebucket-replication.md#s3t-tablebucketreplication-025),
    [S3T-TABLEREPLICATION-035](../table-replication-status.md#s3t-tablereplication-035), [S3T-TABLEBUCKETREPLICATION-015](../tablebucket-replication.md#s3t-tablebucketreplication-015), [S3T-TABLEBUCKETREPLICATION-016](../tablebucket-replication.md#s3t-tablebucketreplication-016), [S3T-TABLE-225](../table-replication-status.md#s3t-table-225),
    [S3T-TABLEBUCKETREPLICATION-005](../tablebucket-replication.md#s3t-tablebucketreplication-005), [S3T-TABLEREPLICATION-015](../table-replication.md#s3t-tablereplication-015), [S3T-TABLEREPLICATION-044](../table-replication.md#s3t-tablereplication-044),
    [S3T-TABLEREPLICATION-016](../table-replication.md#s3t-tablereplication-016), [S3T-TABLEREPLICATION-004](../table-replication.md#s3t-tablereplication-004), [S3T-TABLEREPLICATION-010](../table-replication.md#s3t-tablereplication-010), [S3T-TABLEREPLICATION-021](../table-replication-status.md#s3t-tablereplication-021),
    [S3T-TABLEREPLICATION-022](../table-replication-status.md#s3t-tablereplication-022), [S3T-TABLEBUCKETREPLICATION-013](../tablebucket-replication.md#s3t-tablebucketreplication-013), [S3T-TABLEBUCKETREPLICATION-014](../tablebucket-replication.md#s3t-tablebucketreplication-014) · hypotheses:
    H-U-173, H-U-175 · evidence: table/dependencies/repl-prereqs-token-matrix

## Notes

H-U-173 confirmed (no stale 200). H-U-175 first half confirmed (second Delete -> NotFoundException).

Contradiction with [S3T-TABLEREPLICATION-022](../table-replication-status.md#s3t-tablereplication-022): GetTableReplicationStatus after DeleteTableReplication: 018 ->
NotFoundException at +0.044 s with 0 stale 200 bodies; 022 -> HTTP 200 with the stale table-level view
('D:failed|E:completed') for ~60 s, then the bucket-level view ('D:failed'); the difference is state, not
direction: 022's table was also under bucket-level replication, so the status has a bucket view to fall back
to and only the table-level entries are stale Resolution: keep both; [S3T-TABLEREPLICATION-022](../table-replication-status.md#s3t-tablereplication-022) is the exception
for tables under bucket-level replication, [S3T-TABLEREPLICATION-018](../table-replication.md#s3t-tablereplication-018) holds without a bucket-level config -
018's title is scoped in title_fixes
