<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEREPLICATION-010: First PutTableReplication with versionToken='abc' -> NotFoundException; with '' -> 200; omitted -> 200...
_Full entry and notes of one finding; its summary entry is in [table-replication.md](../table-replication.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-tablereplication-010"></a>**S3T-TABLEREPLICATION-010** `request-validation` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **First PutTableReplication with versionToken='abc' -> NotFoundException; with '' -> 200; omitted -> 200...**
  On a table that never had a replication configuration: PutTableReplication(versionToken='abc') ->
  NotFoundException (HTTP 404) 'The specified replication configuration does not exist.';
  PutTableReplication(versionToken='' , client-side validation disabled) -> HTTP 200; GetTableReplication
  afterwards -> HTTP 200. Omitting versionToken on the first Put -> HTTP 200 (phase 1). Fresh table reads:
  GetTableReplication -> NotFoundException (HTTP 404) 'The specified replication configuration does not
  exist.'; GetTableReplicationStatus -> NotFoundException (HTTP 404) 'The specified replication configuration
  does not exist.'; DeleteTableReplication(versionToken='abc') -> NotFoundException (HTTP 404) 'The specified
  replication configuration does not exist.'.
  - ACK: custom_create, exceptions.404 · ops: PutTableReplication, GetTableReplication,
    GetTableReplicationStatus, DeleteTableReplication · fields: versionToken
  - repro: CreateTable -> Get/Status/Delete('abc') -> Put(versionToken='abc') -> Put(versionToken='') -> Get
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEREPLICATION-001](../table-replication.md#s3t-tablereplication-001), [S3T-TABLEBUCKETREPLICATION-008](../tablebucket-replication.md#s3t-tablebucketreplication-008), [S3T-TABLEBUCKETREPLICATION-005](../tablebucket-replication.md#s3t-tablebucketreplication-005),
    [S3T-TABLEREPLICATION-007](../table-replication.md#s3t-tablereplication-007), [S3T-TABLEBUCKETREPLICATION-015](../tablebucket-replication.md#s3t-tablebucketreplication-015), [S3T-TABLEBUCKETREPLICATION-013](../tablebucket-replication.md#s3t-tablebucketreplication-013),
    [S3T-TABLEREPLICATION-003](../table-replication.md#s3t-tablereplication-003), [S3T-TABLEBUCKETREPLICATION-006](../tablebucket-replication.md#s3t-tablebucketreplication-006), [S3T-TABLEBUCKETREPLICATION-017](../tablebucket-replication.md#s3t-tablebucketreplication-017),
    [S3T-TABLEREPLICATION-004](../table-replication.md#s3t-tablereplication-004), [S3T-TABLEBUCKETREPLICATION-027](../tablebucket-replication.md#s3t-tablebucketreplication-027), [S3T-TABLEREPLICATION-015](../table-replication.md#s3t-tablereplication-015),
    [S3T-TABLEREPLICATION-044](../table-replication.md#s3t-tablereplication-044), [S3T-TABLEREPLICATION-016](../table-replication.md#s3t-tablereplication-016), [S3T-TABLEREPLICATION-018](../table-replication.md#s3t-tablereplication-018) · hypotheses: H-U-191,
    H-U-157, H-U-175 · evidence: table/dependencies/repl-prereqs-token-matrix

## Notes

H-U-191: token 'abc' on a first Put -> NotFoundException (hypothesis said ConflictException 409); empty token
-> accepted (hypothesis said BadRequestException 400). H-U-157 confirmed (both Gets 404 'replication
configuration does not exist', not the table message). H-U-175 second half confirmed: Delete with any token on
a never-configured table -> NotFoundException (existence check precedes the token check).

Full original title: First PutTableReplication with versionToken='abc' -> NotFoundException; with '' -> 200;
omitted -> 200; fresh-table Get/Status/Delete -> NotFoundException
