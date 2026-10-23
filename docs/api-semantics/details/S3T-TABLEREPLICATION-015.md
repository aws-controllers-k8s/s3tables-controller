<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEREPLICATION-015: Stale versionToken: Put x5, Delete x5 -> ConflictException/409 (5/5); Delete('x') -> 409; token-less Put on existing or Delete('') -> 400
_Full entry and notes of one finding; its summary entry is in [table-replication.md](../table-replication.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-tablereplication-015"></a>**S3T-TABLEREPLICATION-015** `idempotency` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Stale versionToken: Put x5, Delete x5 -> ConflictException/409 (5/5); Delete('x') -> 409; token-less Put on existing or Delete('') -> 400**
  With a configuration whose token rotated twice: PutTableReplication(token from 2 Puts ago) five times 2 s
  apart -> ['ConflictException/409', 'ConflictException/409', 'ConflictException/409',
  'ConflictException/409', 'ConflictException/409'] (message 'The specified version token does not match the
  configuration.'); DeleteTableReplication(same stale token) five times -> ['ConflictException/409',
  'ConflictException/409', 'ConflictException/409', 'ConflictException/409', 'ConflictException/409'] (message
  'Provided version token does not match the table version token.'). PutTableReplication WITHOUT versionToken
  while a configuration exists -> BadRequestException (HTTP 400) 'A version token is not specified.'.
  DeleteTableReplication(versionToken='x') -> ConflictException (HTTP 409) 'Provided version token does not
  match the table version token.'; DeleteTableReplication(versionToken='', client validation off) ->
  BadRequestException (HTTP 400) 'A version token is not specified.'. GetTableReplication afterwards -> 200
  (config intact) with the token unchanged.
  - ACK: custom_update, custom_delete, terminal_codes · ops: PutTableReplication, DeleteTableReplication,
    GetTableReplication · fields: versionToken
  - repro: Put (T1) -> Put (T2) -> Put/Delete with T1 x5 @2 s -> Put without token -> Delete('x') ->
    Delete('') -> Get
  - measurements: stale_retries=[5], n=10
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEREPLICATION-004](../table-replication.md#s3t-tablereplication-004), [S3T-TABLEREPLICATION-007](../table-replication.md#s3t-tablereplication-007), [S3T-TABLEBUCKETREPLICATION-026](../tablebucket-replication.md#s3t-tablebucketreplication-026),
    [S3T-TABLEBUCKETREPLICATION-006](../tablebucket-replication.md#s3t-tablebucketreplication-006), [S3T-TABLEBUCKETREPLICATION-004](../tablebucket-replication.md#s3t-tablebucketreplication-004), [S3T-TABLEBUCKETREPLICATION-008](../tablebucket-replication.md#s3t-tablebucketreplication-008),
    [S3T-TABLEREPLICATION-036](../table-replication.md#s3t-tablereplication-036), [S3T-TABLEREPLICATION-044](../table-replication.md#s3t-tablereplication-044), [S3T-TABLEREPLICATION-016](../table-replication.md#s3t-tablereplication-016), [S3T-TABLEREPLICATION-018](../table-replication.md#s3t-tablereplication-018),
    [S3T-TABLEREPLICATION-010](../table-replication.md#s3t-tablereplication-010), [S3T-TABLEREPLICATION-020](../table-replication-status.md#s3t-tablereplication-020), [S3T-TABLEREPLICATION-021](../table-replication-status.md#s3t-tablereplication-021) · hypotheses: H-U-140,
    H-U-141, H-U-142, H-U-174 · evidence: table/dependencies/repl-prereqs-token-matrix

## Notes

H-U-142 confirmed: the stale-token ConflictException is deterministic (5/5 identical for Put, 5/5 for Delete);
'You can retry the request' is misleading - re-read the token instead. H-U-140 (token-less Put =
last-writer-wins) REFUTED and H-U-141 (token-less Put -> 409) REFUTED: it is BadRequestException (HTTP 400) 'A
version token is not specified.'. H-U-174 qualified (garbage token is not a 400): garbage token ->
ConflictException, empty token -> BadRequestException (stale -> ConflictException/409).

Full original title: Stale versionToken: Put x5 over 10 s -> ConflictException/409; Delete x5 ->
ConflictException/409; token-less Put on existing -> BadRequestException; Delete('x') -> ConflictException;
Delete('') -> BadRequestException
