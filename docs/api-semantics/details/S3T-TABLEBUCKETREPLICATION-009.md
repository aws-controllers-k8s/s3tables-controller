<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKETREPLICATION-009: PutTableBucketReplication: duplicate/self/nonexistent destinations -> 200; malformed ARN or >5 destinations -> ValidationException 400
_Full entry and notes of one finding; its summary entry is in
[tablebucket-replication.md](../tablebucket-replication.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucketreplication-009"></a>**S3T-TABLEBUCKETREPLICATION-009** `request-validation` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **PutTableBucketReplication: duplicate/self/nonexistent destinations -> 200; malformed ARN or >5 destinations -> ValidationException 400** (hypothesis refuted; behavior confirmed)
  On an empty source bucket V (role valid, no prior config): destinations [D, D] -> HTTP 200 (Get shows 2
  destination(s)). [V] (self) -> HTTP 200. [V, D] -> HTTP 200 (Get shows 2 destination(s)). A well-formed ARN
  of a bucket that does not exist (same account/region) -> HTTP 200 and Get echoes it. A different-region
  nonexistent bucket -> HTTP 200. Truncated 'arn:aws:s3tables:us-east-1:123:bucket' -> ValidationException
  (HTTP 400) '1 validation error detected: Value at
  'configuration.rules.1.member.destinations.1.member.destinationTableBucketARN' failed to satisfy constraint:
  Member must satisfy regular expression pattern:
  (arn:aws[-a-z0-9]*:[a-z0-9]+:[-a-z0-9]*:[0-9]{12}:bucket/[a-z0-9_-]{3,63})'. Six destinations ->
  ValidationException (HTTP 400) '1 validation error detected: Value at
  'configuration.rules.1.member.destinations' failed to satisfy constraint: Member must have length less than
  or equal to 5'.
  - ACK: terminal_codes, references · ops: PutTableBucketReplication, GetTableBucketReplication · fields:
    configuration.rules[].destinations[].destinationTableBucketARN
  - repro: Fresh bucket; one Put per case (clearing the config in between); Get after each
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-021](../tablebucket-replication.md#s3t-tablebucketreplication-021), [S3T-TABLEBUCKET-011](../tablebucket-replication.md#s3t-tablebucket-011), [S3T-TABLEBUCKETREPLICATION-017](../tablebucket-replication.md#s3t-tablebucketreplication-017),
    [S3T-TABLEREPLICATION-032](../table-replication-status.md#s3t-tablereplication-032), [S3T-TABLEREPLICATION-002](../table-replication-status.md#s3t-tablereplication-002), [S3T-TABLEREPLICATION-008](../table-replication-status.md#s3t-tablereplication-008), [S3T-TABLEREPLICATION-028](../table-replication-status.md#s3t-tablereplication-028),
    [S3T-TABLEBUCKETREPLICATION-022](../tablebucket-replication.md#s3t-tablebucketreplication-022), [S3T-TABLEBUCKETREPLICATION-003](../tablebucket-replication.md#s3t-tablebucketreplication-003), [S3T-TABLEBUCKETREPLICATION-011](../tablebucket-replication.md#s3t-tablebucketreplication-011),
    [S3T-TABLEREPLICATION-017](../table-replication-status.md#s3t-tablereplication-017), [S3T-TABLEBUCKETREPLICATION-010](../tablebucket-replication.md#s3t-tablebucketreplication-010), [S3T-TABLEREPLICATION-005](../table-replication-status.md#s3t-tablereplication-005),
    [S3T-TABLEBUCKETREPLICATION-029](../tablebucket-replication.md#s3t-tablebucketreplication-029), [S3T-TABLEREPLICATION-011](../table-replication.md#s3t-tablereplication-011), [S3T-TABLEREPLICATION-012](../table-replication.md#s3t-tablereplication-012),
    [S3T-TABLEREPLICATION-013](../table-replication.md#s3t-tablereplication-013) · hypotheses: H-S-121, H-S-123, H-S-055 · evidence:
    tablebucket/weird-inputs/repl-validation

## Notes

H-S-121 REFUTED: [D, D] accepted and stored twice (no 400, no dedup). H-S-123 REFUTED: the source bucket
itself as the only destination is accepted (200) and stored; [self, D] stored with 2 entries. H-S-055
qualified: nonexistent destination accepted (as hypothesised) but self/duplicate are NOT rejected; a malformed
ARN is rejected with code ValidationException (not the modeled BadRequestException) and HTTP 400, and more
than 5 destinations is rejected the same way. A CRD must pre-validate uniqueness/self-reference itself; the
service will not.

Full original title: PutTableBucketReplication validates only ARN syntax and list length (ValidationException
400); duplicate, self, nonexistent, foreign destinations all 200
