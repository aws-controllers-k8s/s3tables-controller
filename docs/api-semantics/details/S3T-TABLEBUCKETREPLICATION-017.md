<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKETREPLICATION-017: PutTableBucketReplication rules/destinations bounds -> ValidationException 400 (code not in the model); 5 dests + non-existent role -> 200
_Full entry and notes of one finding; its summary entry is in
[tablebucket-replication.md](../tablebucket-replication.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucketreplication-017"></a>**S3T-TABLEBUCKETREPLICATION-017** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **PutTableBucketReplication rules/destinations bounds -> ValidationException 400 (code not in the model); 5 dests + non-existent role -> 200**
  PutTableBucketReplication on a fresh table bucket, role = a syntactically valid but NON-EXISTENT IAM role
  ARN, client validation off: 6_destinations -> ValidationException (HTTP 400) '1 validation error detected:
  Value at 'configuration.rules.1.member.destinations' failed to satisfy constraint: Member must have length
  less than or equal to 5'; 2_rules -> ValidationException (HTTP 400) '1 validation error detected: Value at
  'configuration.rules' failed to satisfy constraint: Member must have length less than or equal to 1';
  0_rules -> ValidationException (HTTP 400) '1 validation error detected: Value at 'configuration.rules'
  failed to satisfy constraint: Member must have length greater than or equal to 1'; rule_with_0_destinations
  -> ValidationException (HTTP 400) '1 validation error detected: Value at
  'configuration.rules.1.member.destinations' failed to satisfy constraint: Member must have length greater
  than or equal to 1'; 5_destinations -> HTTP 200; 1_destination -> BadRequestException (HTTP 400) 'A version
  token is not specified.'. The list-size checks fire first and use the wire code 'ValidationException' (HTTP
  400), which is NOT among the modeled errors (BadRequestException is). The 5-destination configuration with a
  role ARN that does not exist was ACCEPTED (HTTP 200). A subsequent DeleteTableBucketReplication and
  PutTableBucketReplication without versionToken were rejected with BadRequestException 400 'A version token
  is not specified.'; the bucket with the lingering replication configuration was still deletable with
  DeleteTableBucket.
  - ACK: terminal_codes, none · ops: PutTableBucketReplication · fields: configuration.rules,
    configuration.rules[].destinations, configuration.role
  - repro: PutTableBucketReplication(configuration={role: <nonexistent>, rules: <6 dest | 2 rules | [] |
    [{destinations: []}] | 5 dest>})
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-019](../tablebucket-replication.md#s3t-tablebucketreplication-019), [S3T-TABLEBUCKETREPLICATION-020](../tablebucket-replication.md#s3t-tablebucketreplication-020), [S3T-TABLEBUCKETREPLICATION-001](../tablebucket-replication.md#s3t-tablebucketreplication-001),
    [S3T-TABLEBUCKETREPLICATION-002](../tablebucket-replication.md#s3t-tablebucketreplication-002), [S3T-TABLEBUCKETREPLICATION-003](../tablebucket-replication.md#s3t-tablebucketreplication-003), [S3T-TABLEBUCKETREPLICATION-004](../tablebucket-replication.md#s3t-tablebucketreplication-004),
    [S3T-TABLEBUCKETREPLICATION-005](../tablebucket-replication.md#s3t-tablebucketreplication-005), [S3T-TABLEBUCKETREPLICATION-021](../tablebucket-replication.md#s3t-tablebucketreplication-021), [S3T-TABLEBUCKET-011](../tablebucket-replication.md#s3t-tablebucket-011),
    [S3T-TABLEBUCKETREPLICATION-009](../tablebucket-replication.md#s3t-tablebucketreplication-009), [S3T-TABLEREPLICATION-032](../table-replication-status.md#s3t-tablereplication-032), [S3T-TABLEREPLICATION-002](../table-replication-status.md#s3t-tablereplication-002),
    [S3T-TABLEREPLICATION-008](../table-replication-status.md#s3t-tablereplication-008), [S3T-TABLEREPLICATION-028](../table-replication-status.md#s3t-tablereplication-028), [S3T-TABLEBUCKETREPLICATION-006](../tablebucket-replication.md#s3t-tablebucketreplication-006),
    [S3T-TABLEREPLICATION-004](../table-replication.md#s3t-tablereplication-004), [S3T-TABLEREPLICATION-007](../table-replication.md#s3t-tablereplication-007), [S3T-TABLEREPLICATION-010](../table-replication.md#s3t-tablereplication-010),
    [S3T-TABLEBUCKETREPLICATION-027](../tablebucket-replication.md#s3t-tablebucketreplication-027), [S3T-TABLEBUCKETREPLICATION-011](../tablebucket-replication.md#s3t-tablebucketreplication-011), [S3T-TABLEBUCKETREPLICATION-010](../tablebucket-replication.md#s3t-tablebucketreplication-010),
    [S3T-TABLEREPLICATION-005](../table-replication-status.md#s3t-tablereplication-005), [S3T-TABLEBUCKETREPLICATION-029](../tablebucket-replication.md#s3t-tablebucketreplication-029), [S3T-TABLEBUCKETREPLICATION-012](../tablebucket-replication.md#s3t-tablebucketreplication-012),
    [S3T-TABLEBUCKETREPLICATION-028](../tablebucket-replication.md#s3t-tablebucketreplication-028), [S3T-TABLEREPLICATION-009](../table-replication.md#s3t-tablereplication-009) · hypotheses: H-S-056 · evidence:
    tablebucket/limits/bucket-quota-create-burst

## Notes

H-S-056 qualified: the limits exist (>5 destinations, >1 rule, [] rules, [] destinations all rejected) but as
ValidationException, not BadRequestException; 5 destinations accepted. The IAM role was not validated at Put
time (non-existent role ARN accepted). No IAM role was created (shard: replication probes hold
s3tables:replication). The SDK model bounds rules to 1..1 and destinations to 1..5, so with client-side
validation on, botocore rejects only the min violations; max violations reach AWS.
