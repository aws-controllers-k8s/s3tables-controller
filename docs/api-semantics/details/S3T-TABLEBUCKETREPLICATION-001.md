<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKETREPLICATION-001: Bucket replication token: 20 Gets identical; garbage-token Put -> ConflictException (token unchanged)...
_Full entry and notes of one finding; its summary entry is in
[tablebucket-replication.md](../tablebucket-replication.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucketreplication-001"></a>**S3T-TABLEBUCKETREPLICATION-001** `idempotency` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Bucket replication token: 20 Gets identical; garbage-token Put -> ConflictException (token unchanged)...**
  After PutTableBucketReplication (status='success') GetTableBucketReplication x20 at 1/s returned one
  identical versionToken (len 20, equal to the Put's token). Put with versionToken='abc' -> ConflictException
  (HTTP 409) 'The specified version token does not match the configuration.'; the Get token afterwards was
  unchanged. Put of a DIFFERENT configuration WITHOUT versionToken while a configuration exists ->
  BadRequestException (HTTP 400) 'A version token is not specified.'; Get then showed destinations [D] (old
  config kept) and the token unchanged.
  - ACK: custom_update, requeue · ops: PutTableBucketReplication, GetTableBucketReplication · fields:
    versionToken
  - repro: Put -> Get x20 -> Put(versionToken='abc') -> Get -> Put(other config, no token) -> Get
  - measurements: get_x20_distinct_tokens=[1], token_len=[20], n=20
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-018](../tablebucket-replication.md#s3t-tablebucketreplication-018), [S3T-TABLEBUCKETREPLICATION-003](../tablebucket-replication.md#s3t-tablebucketreplication-003), [S3T-TABLEBUCKETREPLICATION-007](../tablebucket-replication.md#s3t-tablebucketreplication-007),
    [S3T-TABLEREPLICATION-037](../table-replication.md#s3t-tablereplication-037), [S3T-TABLEREPLICATION-003](../table-replication.md#s3t-tablereplication-003), [S3T-TABLEREPLICATION-001](../table-replication.md#s3t-tablereplication-001),
    [S3T-TABLEBUCKETREPLICATION-019](../tablebucket-replication.md#s3t-tablebucketreplication-019), [S3T-TABLEBUCKETREPLICATION-020](../tablebucket-replication.md#s3t-tablebucketreplication-020), [S3T-TABLEBUCKETREPLICATION-002](../tablebucket-replication.md#s3t-tablebucketreplication-002),
    [S3T-TABLEBUCKETREPLICATION-017](../tablebucket-replication.md#s3t-tablebucketreplication-017), [S3T-TABLEBUCKETREPLICATION-004](../tablebucket-replication.md#s3t-tablebucketreplication-004), [S3T-TABLEBUCKETREPLICATION-005](../tablebucket-replication.md#s3t-tablebucketreplication-005),
    [S3T-TABLEBUCKET-012](../tablebucket-replication.md#s3t-tablebucket-012), [S3T-SERVICE-016](../service.md#s3t-service-016) · hypotheses: H-S-119, H-S-050 · evidence:
    tablebucket/dependencies/repl-setup

## Notes

H-S-119 confirmed (20 Gets identical, failed Put does not move the token). H-S-050 confirmed in its 400
variant: a token-less Put on an existing configuration is rejected with BadRequestException 400 'A version
token is not specified.' (not ConflictException 409); the stored configuration and token are unchanged.

Full original title: Bucket replication token: 20 Gets identical; garbage-token Put -> ConflictException
(token unchanged); token-less Put on existing config -> BadRequestException
