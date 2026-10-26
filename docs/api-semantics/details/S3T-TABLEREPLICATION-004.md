<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEREPLICATION-004: Replication versionToken != GetTable.versionToken; Put with the table token -> ConflictException...
_Full entry and notes of one finding; its summary entry is in [table-replication.md](../table-replication.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-tablereplication-004"></a>**S3T-TABLEREPLICATION-004** `identity` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Replication versionToken != GetTable.versionToken; Put with the table token -> ConflictException...**
  After PutTableReplication the table's own GetTable.versionToken did not change (it differs from the
  replication token). PutTableReplication with versionToken=GetTable's token -> ConflictException (HTTP 409)
  'The specified version token does not match the configuration.'. PutTableReplication without a token while a
  configuration exists -> BadRequestException (HTTP 400) 'A version token is not specified.'. Re-Put of the
  identical configuration with the current replication token -> HTTP 200, token rotated.
  DeleteTableReplication with GetTable's token -> ConflictException (HTTP 409) 'Provided version token does
  not match the table version token.'.
  - ACK: custom_update, custom_delete · ops: PutTableReplication, GetTable, DeleteTableReplication · fields:
    versionToken
  - repro: PutTableReplication -> GetTable -> Put(versionToken=GetTable token) -> Put(no token) ->
    Put(current) -> DeleteTableReplication(GetTable token)
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-006](../tablebucket-replication.md#s3t-tablebucketreplication-006), [S3T-TABLEBUCKETREPLICATION-017](../tablebucket-replication.md#s3t-tablebucketreplication-017), [S3T-TABLEREPLICATION-007](../table-replication.md#s3t-tablereplication-007),
    [S3T-TABLEBUCKETREPLICATION-005](../tablebucket-replication.md#s3t-tablebucketreplication-005), [S3T-TABLEREPLICATION-010](../table-replication.md#s3t-tablereplication-010), [S3T-TABLEBUCKETREPLICATION-027](../tablebucket-replication.md#s3t-tablebucketreplication-027),
    [S3T-TABLEREPLICATION-015](../table-replication.md#s3t-tablereplication-015), [S3T-TABLEREPLICATION-044](../table-replication.md#s3t-tablereplication-044), [S3T-TABLEREPLICATION-016](../table-replication.md#s3t-tablereplication-016), [S3T-TABLEREPLICATION-018](../table-replication.md#s3t-tablereplication-018) ·
    hypotheses: H-U-137 · evidence: table/round-trip/repl-put-get

## Notes

H-U-137 confirmed (distinct tokens; the table token is rejected with ConflictException). Table-level token
rules mirror the bucket level: token-less Put on existing -> BadRequestException, unchanged re-Put rotates.

Full original title: Replication versionToken != GetTable.versionToken; Put with the table token ->
ConflictException; token-less Put on existing -> BadRequestException; Delete with table token ->
ConflictException
