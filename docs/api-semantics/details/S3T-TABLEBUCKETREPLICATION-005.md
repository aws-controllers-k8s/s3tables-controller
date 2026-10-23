<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKETREPLICATION-005: After DeleteTableBucketReplication: Get -> 404; Put with the pre-delete token -> 404 (not 409); token-less Put -> 200 with a new token
_Full entry and notes of one finding; its summary entry is in
[tablebucket-replication.md](../tablebucket-replication.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucketreplication-005"></a>**S3T-TABLEBUCKETREPLICATION-005** `error-code` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **After DeleteTableBucketReplication: Get -> 404; Put with the pre-delete token -> 404 (not 409); token-less Put -> 200 with a new token**
  After deleting an existing configuration with its current token: GetTableBucketReplication ->
  NotFoundException (HTTP 404) 'The specified replication configuration does not exist.' (error body carries
  no versionToken). PutTableBucketReplication WITH the token that was current before the Delete ->
  NotFoundException (HTTP 404) 'The specified replication configuration does not exist.' (a 404, not a 409).
  PutTableBucketReplication WITHOUT a token -> HTTP 200 with a versionToken different from every earlier
  token.
  - ACK: exceptions.404, custom_delete · ops: GetTableBucketReplication, DeleteTableBucketReplication,
    PutTableBucketReplication · fields: versionToken
  - repro: Put -> Delete(current token) -> Get -> Put(old token) -> Put(no token)
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-023](../tablebucket-replication.md#s3t-tablebucketreplication-023), [S3T-TABLEBUCKETREPLICATION-008](../tablebucket-replication.md#s3t-tablebucketreplication-008), [S3T-TABLEREPLICATION-034](../table-replication.md#s3t-tablereplication-034),
    [S3T-TABLEREPLICATION-001](../table-replication.md#s3t-tablereplication-001), [S3T-TABLEREPLICATION-021](../table-replication-status.md#s3t-tablereplication-021), [S3T-TABLEREPLICATION-037](../table-replication.md#s3t-tablereplication-037),
    [S3T-TABLEBUCKETREPLICATION-019](../tablebucket-replication.md#s3t-tablebucketreplication-019), [S3T-TABLEBUCKETREPLICATION-020](../tablebucket-replication.md#s3t-tablebucketreplication-020), [S3T-TABLEBUCKETREPLICATION-001](../tablebucket-replication.md#s3t-tablebucketreplication-001),
    [S3T-TABLEBUCKETREPLICATION-002](../tablebucket-replication.md#s3t-tablebucketreplication-002), [S3T-TABLEBUCKETREPLICATION-003](../tablebucket-replication.md#s3t-tablebucketreplication-003), [S3T-TABLEBUCKETREPLICATION-017](../tablebucket-replication.md#s3t-tablebucketreplication-017),
    [S3T-TABLEBUCKETREPLICATION-004](../tablebucket-replication.md#s3t-tablebucketreplication-004), [S3T-TABLEBUCKETREPLICATION-024](../tablebucket-replication.md#s3t-tablebucketreplication-024), [S3T-TABLEBUCKETREPLICATION-025](../tablebucket-replication.md#s3t-tablebucketreplication-025),
    [S3T-TABLEREPLICATION-035](../table-replication-status.md#s3t-tablereplication-035), [S3T-TABLEBUCKETREPLICATION-015](../tablebucket-replication.md#s3t-tablebucketreplication-015), [S3T-TABLEBUCKETREPLICATION-016](../tablebucket-replication.md#s3t-tablebucketreplication-016),
    [S3T-TABLEREPLICATION-018](../table-replication.md#s3t-tablereplication-018), [S3T-TABLE-225](../table-replication-status.md#s3t-table-225), [S3T-TABLEREPLICATION-010](../table-replication.md#s3t-tablereplication-010), [S3T-TABLEREPLICATION-007](../table-replication.md#s3t-tablereplication-007),
    [S3T-TABLEBUCKETREPLICATION-013](../tablebucket-replication.md#s3t-tablebucketreplication-013), [S3T-TABLEREPLICATION-003](../table-replication.md#s3t-tablereplication-003), [S3T-TABLEBUCKETREPLICATION-006](../tablebucket-replication.md#s3t-tablebucketreplication-006),
    [S3T-TABLEREPLICATION-004](../table-replication.md#s3t-tablereplication-004), [S3T-TABLEBUCKETREPLICATION-027](../tablebucket-replication.md#s3t-tablebucketreplication-027), [S3T-TABLEBUCKETREPLICATION-032](../tablebucket-replication.md#s3t-tablebucketreplication-032), [S3T-TABLE-138](../table-replication-status.md#s3t-table-138),
    [S3T-TABLE-196](../table-replication-status.md#s3t-table-196) · hypotheses: H-S-120 · evidence: tablebucket/dependencies/repl-setup

## Notes

H-S-120 confirmed, with one precision: sending the old token after a delete yields NotFoundException 404 (not
409/400), so a controller that sees 404 on Put must clear its token and re-Put without one. The
never-configured half of H-S-047 is probed in tablebucket/weird-inputs/repl-validation (in this run the
comparison bucket still carried a configuration because of a script bug).

Full original title: After DeleteTableBucketReplication: Get -> 404 'does not exist', Put with the pre-delete
token -> 404 (not 409), token-less Put -> 200 new lineage
