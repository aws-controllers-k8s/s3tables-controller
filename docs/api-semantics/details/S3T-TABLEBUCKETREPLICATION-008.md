<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKETREPLICATION-008: Never-configured bucket: GetTableBucketReplication -> 404; DeleteTableBucketReplication -> 404 with or without a token (never 400)
_Full entry and notes of one finding; its summary entry is in
[tablebucket-replication.md](../tablebucket-replication.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucketreplication-008"></a>**S3T-TABLEBUCKETREPLICATION-008** `error-code` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Never-configured bucket: GetTableBucketReplication -> 404; DeleteTableBucketReplication -> 404 with or without a token (never 400)**
  On a fresh table bucket that never had a replication configuration: GetTableBucketReplication ->
  NotFoundException (HTTP 404) 'The specified replication configuration does not exist.';
  DeleteTableBucketReplication without versionToken -> NotFoundException (HTTP 404) 'The specified replication
  configuration does not exist.'; DeleteTableBucketReplication with versionToken='abc' -> NotFoundException
  (HTTP 404) 'The specified replication configuration does not exist.' (repeated: NotFoundException (HTTP 404)
  'The specified replication configuration does not exist.').
  - ACK: exceptions.404, custom_delete · ops: GetTableBucketReplication, DeleteTableBucketReplication ·
    fields: versionToken
  - repro: CreateTableBucket -> GetTableBucketReplication; DeleteTableBucketReplication (no token / token
    'abc') x2
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-023](../tablebucket-replication.md#s3t-tablebucketreplication-023), [S3T-TABLEBUCKETREPLICATION-005](../tablebucket-replication.md#s3t-tablebucketreplication-005), [S3T-TABLEREPLICATION-034](../table-replication.md#s3t-tablereplication-034),
    [S3T-TABLEREPLICATION-001](../table-replication.md#s3t-tablereplication-001), [S3T-TABLEREPLICATION-021](../table-replication-status.md#s3t-tablereplication-021), [S3T-TABLEREPLICATION-037](../table-replication.md#s3t-tablereplication-037),
    [S3T-TABLEBUCKETREPLICATION-026](../tablebucket-replication.md#s3t-tablebucketreplication-026), [S3T-TABLEBUCKETREPLICATION-006](../tablebucket-replication.md#s3t-tablebucketreplication-006), [S3T-TABLEBUCKETREPLICATION-004](../tablebucket-replication.md#s3t-tablebucketreplication-004),
    [S3T-TABLEREPLICATION-036](../table-replication.md#s3t-tablereplication-036), [S3T-TABLEREPLICATION-007](../table-replication.md#s3t-tablereplication-007), [S3T-TABLEREPLICATION-015](../table-replication.md#s3t-tablereplication-015), [S3T-TABLEREPLICATION-010](../table-replication.md#s3t-tablereplication-010),
    [S3T-TABLEBUCKETREPLICATION-015](../tablebucket-replication.md#s3t-tablebucketreplication-015), [S3T-TABLEBUCKETREPLICATION-013](../tablebucket-replication.md#s3t-tablebucketreplication-013), [S3T-TABLEREPLICATION-003](../table-replication.md#s3t-tablereplication-003) · hypotheses:
    H-S-047 · evidence: tablebucket/weird-inputs/repl-validation

## Notes

H-S-047 confirmed: Get -> NotFoundException 404 and Delete-when-none -> NotFoundException 404 (not an
idempotent 204), both with and without a token. Note the asymmetry with a CONFIGURED bucket, where a
token-less Delete is rejected with BadRequestException 400 'A version token is not specified.'
([S3T-TABLEBUCKETREPLICATION-006](../tablebucket-replication.md#s3t-tablebucketreplication-006)): the existence check runs before the token check.
