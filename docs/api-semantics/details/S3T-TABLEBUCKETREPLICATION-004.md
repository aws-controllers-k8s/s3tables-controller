<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKETREPLICATION-004: DeleteTableBucketReplication: stale token -> ConflictException, garbage token -> ConflictException, current token -> HTTP 204...
_Full entry and notes of one finding; its summary entry is in
[tablebucket-replication.md](../tablebucket-replication.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucketreplication-004"></a>**S3T-TABLEBUCKETREPLICATION-004** `delete-semantics` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **DeleteTableBucketReplication: stale token -> ConflictException, garbage token -> ConflictException, current token -> HTTP 204...**
  With a configuration whose token had rotated twice: Delete(versionToken=<token from 2 Puts ago>) ->
  ConflictException (HTTP 409) 'The specified version token does not match the configuration.' (config still
  present); Delete(versionToken='abc') -> ConflictException (HTTP 409) 'The specified version token does not
  match the configuration.' (config still present); Delete(current token) -> HTTP 204;
  GetTableBucketReplication polled at 200 ms returned NotFoundException on the first poll (0.01s).
  - ACK: custom_delete, exceptions.404 · ops: DeleteTableBucketReplication, GetTableBucketReplication ·
    fields: versionToken
  - repro: Put (T1) -> Put (T2) -> Delete(T1) -> Get -> Delete('abc') -> Get -> Delete(T2) -> Get
  - measurements: get_404_after_delete_s=[0.01], n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-019](../tablebucket-replication.md#s3t-tablebucketreplication-019), [S3T-TABLEBUCKETREPLICATION-020](../tablebucket-replication.md#s3t-tablebucketreplication-020), [S3T-TABLEBUCKETREPLICATION-001](../tablebucket-replication.md#s3t-tablebucketreplication-001),
    [S3T-TABLEBUCKETREPLICATION-002](../tablebucket-replication.md#s3t-tablebucketreplication-002), [S3T-TABLEBUCKETREPLICATION-003](../tablebucket-replication.md#s3t-tablebucketreplication-003), [S3T-TABLEBUCKETREPLICATION-017](../tablebucket-replication.md#s3t-tablebucketreplication-017),
    [S3T-TABLEBUCKETREPLICATION-005](../tablebucket-replication.md#s3t-tablebucketreplication-005), [S3T-TABLEBUCKETREPLICATION-026](../tablebucket-replication.md#s3t-tablebucketreplication-026), [S3T-TABLEBUCKETREPLICATION-006](../tablebucket-replication.md#s3t-tablebucketreplication-006),
    [S3T-TABLEBUCKETREPLICATION-008](../tablebucket-replication.md#s3t-tablebucketreplication-008), [S3T-TABLEREPLICATION-036](../table-replication.md#s3t-tablereplication-036), [S3T-TABLEREPLICATION-007](../table-replication.md#s3t-tablereplication-007),
    [S3T-TABLEREPLICATION-015](../table-replication.md#s3t-tablereplication-015), [S3T-TABLEBUCKET-012](../tablebucket-replication.md#s3t-tablebucket-012), [S3T-SERVICE-016](../service.md#s3t-service-016) · hypotheses: H-S-129 · evidence:
    tablebucket/dependencies/repl-setup

## Notes

H-S-129 qualified: a stale token -> ConflictException 409 and the configuration remains (confirmed); a
syntactically garbage token 'abc' -> ConflictException 409 as well, NOT BadRequestException 400 (refuted for
that half). Only the current token deletes (a token-less Delete is rejected with 400, see the H-S-051
finding).

Full original title: DeleteTableBucketReplication: stale token -> ConflictException, garbage token ->
ConflictException, current token -> HTTP 204; Get 404 at +0.01s
