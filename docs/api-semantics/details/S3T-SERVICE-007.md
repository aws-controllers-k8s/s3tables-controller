<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-SERVICE-007: PutTableReplication/PutTableBucketReplication with body {} -> InternalServerErrorException 500 in 6/6 sends (deterministic, no side effect)
_Full entry and notes of one finding; its summary entry is in [service.md](../service.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-service-007"></a>**S3T-SERVICE-007** `error-code` · impact medium · handled · verified 2026-10-14
  **PutTableReplication/PutTableBucketReplication with body {} -> InternalServerErrorException 500 in 6/6 sends (deterministic, no side effect)**
  Raw signed PUT with body '{}' (no 'configuration' member): PutTableReplication {} #1 -> HTTP 500
  InternalServerErrorException "{\"message\":\"An internal error occurred. Try again.\"}";
  PutTableBucketReplication {} #1 -> HTTP 500 InternalServerErrorException "{\"message\":\"An internal error
  occurred. Try again.\"}"; PutTableReplication {} #2 -> HTTP 500 InternalServerErrorException
  "{\"message\":\"An internal error occurred. Try again.\"}"; PutTableBucketReplication {} #2 -> HTTP 500
  InternalServerErrorException "{\"message\":\"An internal error occurred. Try again.\"}"; PutTableReplication
  {} #3 -> HTTP 500 InternalServerErrorException "{\"message\":\"An internal error occurred. Try again.\"}";
  PutTableBucketReplication {} #3 -> HTTP 500 InternalServerErrorException "{\"message\":\"An internal error
  occurred. Try again.\"}"; PutTableReplication {configuration:{}} -> HTTP 400 BadRequestException
  "{\"message\":\"Invalid input: Must provide either table arn alone OR all of (table bucket arn, namespace,
  tab; PutTableReplication {configuration:null} -> HTTP 500 InternalServerErrorException "{\"message\":\"An
  internal error occurred. Try again.\"}"; PutTableReplication {configuration:{rules:[]}} -> HTTP 400
  BadRequestException "{\"message\":\"Invalid input: Must provide either table arn alone OR all of (table
  bucket arn, namespace, tab; PutTableReplication {configuration:{role:x}} (rules omitted) -> HTTP 403
  ForbiddenException "{\"message\":\"Cross-account pass role is not allowed.\"}"; PutTableReplication {} no
  tableArn query -> HTTP 400 BadRequestException "{\"message\":\"A table ARN is not specified\"}";
  PutTableReplication {} ghost table uuid -> HTTP 500 InternalServerErrorException "{\"message\":\"An internal
  error occurred. Try again.\"}"; PutTableReplication {} bucket ARN as tableArn -> HTTP 400
  BadRequestException "{\"message\":\"The specified table ARN is not valid.\"}"; PutTableReplication
  {versionToken:abc} only -> HTTP 500 InternalServerErrorException "{\"message\":\"An internal error occurred.
  Try again.\"}"; PutTableBucketReplication {configuration:{}} -> HTTP 400 BadRequestException
  "{\"message\":\"Invalid input: Must provide either table arn alone OR all of (table bucket arn, namespace,
  tab; PutTableBucketReplication {configuration:null} -> HTTP 500 InternalServerErrorException
  "{\"message\":\"An internal error occurred. Try again.\"}"; PutTableBucketReplication {} no query -> HTTP
  400 BadRequestException "{\"message\":\"The specified table bucket ARN is not valid.\"}";
  PutTableBucketReplication {} ghost bucket -> HTTP 500 InternalServerErrorException "{\"message\":\"An
  internal error occurred. Try again.\"}"; PutTableBucketReplication {} table ARN as bucket -> HTTP 400
  BadRequestException "{\"message\":\"The specified bucket name isn't valid. Specify a different bucket name,
  and then try again.\"}; PutTableBucketReplication {versionToken:abc} only -> HTTP 500
  InternalServerErrorException "{\"message\":\"An internal error occurred. Try again.\"}";
  PutTableRecordExpirationConfiguration {} (control, ValidationException expected) -> HTTP 400
  ValidationException "{\"message\":\"1 validation error detected: Value at 'value' failed to satisfy
  constraint: Member must not be. Side effects: GetTableReplication -> NotFoundException,
  GetTableBucketReplication -> NotFoundException, GetTable -> 200.
  - ACK: terminal_codes, requeue · ops: PutTableReplication, PutTableBucketReplication · fields: configuration
  - repro: SigV4 PUT /table-replication?tableArn=<arn> body '{}' x3; PUT
    /table-bucket-replication?tableBucketARN=<arn> body '{}' x3; variants with configuration {} / null
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-SERVICE-002](../service.md#s3t-service-002), [S3T-SERVICE-020](../service.md#s3t-service-020), [S3T-SERVICE-026](../service.md#s3t-service-026), [S3T-TABLEBUCKETREPLICATION-003](../tablebucket-replication.md#s3t-tablebucketreplication-003),
    [S3T-SERVICE-015](../service.md#s3t-service-015), [S3T-SERVICE-021](../service.md#s3t-service-021), [S3T-TABLE-300](../table-replication-status.md#s3t-table-300), [S3T-TABLE-304](../table-replication.md#s3t-table-304) · evidence:
    service/creative/sig-skew-and-empty-500

## Notes

Creative round, seed error-messages: a request-shape error (missing required 'configuration', which the SDK
would reject client-side) surfaces as a 500 'An internal error occurred. Try again.' - PERMANENT-SPEC behind a
RETRY code. Variant bodies in behavior show which member triggers it (configuration absent vs {} vs null).
500-returning variants: ['PutTableBucketReplication {configuration:null}', 'PutTableBucketReplication
{versionToken:abc} only', 'PutTableBucketReplication {} #1', 'PutTableBucketReplication {} #2',
'PutTableBucketReplication {} #3', 'PutTableBucketReplication {} ghost bucket', 'PutTableReplication
{configuration:null}', 'PutTableReplication {versionToken:abc} only', 'PutTableReplication {} #1',
'PutTableReplication {} #2', 'PutTableReplication {} #3', 'PutTableReplication {} ghost table uuid'].
