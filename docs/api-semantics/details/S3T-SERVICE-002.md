<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-SERVICE-002: Empty body {} on 16 write ops: 6 ValidationException (framework), 7 BadRequestException (handler), 3 other - two validation layers
_Full entry and notes of one finding; its summary entry is in [service.md](../service.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-service-002"></a>**S3T-SERVICE-002** `error-code` · impact medium · handled · verified 2026-10-14
  **Empty body {} on 16 write ops: 6 ValidationException (framework), 7 BadRequestException (handler), 3 other - two validation layers**
  Signed raw request with body '{}' and no other members: CreateTableBucket -> HTTP 400 BadRequestException
  "{\"message\":\"The specified bucket name isn't valid. Specify a different bucket name, and then try
  again.\"}"; CreateNamespace -> HTTP 400 BadRequestException "{\"message\":\"The specified namespace name
  isn\u2019t valid. Specify a different namespace name, and then try again.\"; CreateTable -> HTTP 400
  BadRequestException "{\"message\":\"The specified open table format is not valid\"}"; PutTableBucketPolicy
  -> HTTP 400 ValidationException "{\"message\":\"1 validation error detected: Value at 'resourcePolicy'
  failed to satisfy constraint: Member must not be ; PutTableBucketEncryption -> HTTP 400 BadRequestException
  "{\"message\":\"An encryption configuration is not specified\"}"; PutTableBucketStorageClass -> HTTP 400
  ValidationException "{\"message\":\"1 validation error detected: Value at 'storageClassConfiguration' failed
  to satisfy constraint: Member m; PutTableBucketMaintenanceConfiguration -> HTTP 400 ValidationException
  "{\"message\":\"1 validation error detected: Value at 'value' failed to satisfy constraint: Member must not
  be null\"}"; PutTableBucketMetricsConfiguration -> HTTP 204 - ""; PutTableBucketReplication -> HTTP 500
  InternalServerErrorException "{\"message\":\"An internal error occurred. Try again.\"}";
  PutTableMaintenanceConfiguration -> HTTP 400 BadRequestException "{\"message\":\"The table maintenance
  configuration value is not specified.\"}"; UpdateTableMetadataLocation -> HTTP 400 ValidationException
  "{\"message\":\"2 validation errors detected: Value at 'versionToken' failed to satisfy constraint: Member
  must not be n; RenameTable -> HTTP 400 BadRequestException "{\"message\":\"Neither a new namespace name nor
  a new table name is specified.\"}"; PutTablePolicy -> HTTP 400 ValidationException "{\"message\":\"1
  validation error detected: Value at 'resourcePolicy' failed to satisfy constraint: Member must not be ;
  PutTableReplication -> HTTP 500 InternalServerErrorException "{\"message\":\"An internal error occurred. Try
  again.\"}"; PutTableRecordExpirationConfiguration -> HTTP 400 ValidationException "{\"message\":\"1
  validation error detected: Value at 'value' failed to satisfy constraint: Member must not be null\"}";
  TagResource -> HTTP 400 BadRequestException "{\"message\":\"The provided tags are invalid.\"}".
  - ACK: terminal_codes, docs-only · ops: CreateTableBucket, CreateNamespace, CreateTable,
    PutTableBucketPolicy, PutTableBucketEncryption, PutTableBucketStorageClass,
    PutTableBucketMaintenanceConfiguration, PutTableBucketMetricsConfiguration, PutTableBucketReplication,
    PutTableMaintenanceConfiguration, UpdateTableMetadataLocation, RenameTable, PutTablePolicy,
    PutTableReplication, PutTableRecordExpirationConfiguration, TagResource
  - repro: For each write op: SigV4 PUT/POST <modeled path> with body '{}' (content-type application/json)
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-SERVICE-007](../service.md#s3t-service-007), [S3T-SERVICE-020](../service.md#s3t-service-020) · evidence: service/creative/raw-wire-codes

## Notes

ValidationException ops: ['PutTableBucketMaintenanceConfiguration', 'PutTableBucketPolicy',
'PutTableBucketStorageClass', 'PutTablePolicy', 'PutTableRecordExpirationConfiguration',
'UpdateTableMetadataLocation']. BadRequestException ops: ['CreateNamespace', 'CreateTable',
'CreateTableBucket', 'PutTableBucketEncryption', 'PutTableMaintenanceConfiguration', 'RenameTable',
'TagResource']. Other: ['PutTableBucketMetricsConfiguration:204/None',
'PutTableBucketReplication:500/InternalServerErrorException',
'PutTableReplication:500/InternalServerErrorException']. Both are HTTP 400 and both are PERMANENT-SPEC;
ValidationException is not in the s3tables model (no SDK type) so a generated controller sees it as an
unmodeled error and must match on the wire code string. Malformed/typed-body probes: CreateTableBucket
malformed '{"name":' -> HTTP 400 BadRequestException "{\"message\":\"Invalid input. Consult the S3 Tables
and/or Iceberg REST Catalog API reference for constraints on reques; PutTablePolicy malformed '{' -> HTTP 400
BadRequestException "{\"message\":\"Invalid input. Consult the S3 Tables and/or Iceberg REST Catalog API
reference for constraints on reques; CreateTableBucket name:123 (int for string) -> HTTP 400
BadRequestException "{\"message\":\"Invalid input. Consult the S3 Tables and/or Iceberg REST Catalog API
reference for constraints on reques; CreateTableBucket name:null -> HTTP 400 BadRequestException
"{\"message\":\"The specified bucket name isn't valid. Specify a different bucket name, and then try
again.\"}"; CreateNamespace namespace:'str' (string for list) -> HTTP 400 BadRequestException
"{\"message\":\"Invalid input. Consult the S3 Tables and/or Iceberg REST Catalog API reference for constraints
on reques; PutTableBucketStorageClass storageClass:5 -> HTTP 400 BadRequestException "{\"message\":\"Invalid
input. Consult the S3 Tables and/or Iceberg REST Catalog API reference for constraints on reques; TagResource
tags:[] (list for map) -> HTTP 400 BadRequestException "{\"message\":\"Invalid input. Consult the S3 Tables
and/or Iceberg REST Catalog API reference for constraints on reques; CreateTableBucket empty body '' -> HTTP
400 BadRequestException "{\"message\":\"The specified bucket name isn't valid. Specify a different bucket
name, and then try again.\"}"; CreateTableBucket body 'null' -> HTTP 400 BadRequestException
"{\"message\":\"Invalid input. Consult the S3 Tables and/or Iceberg REST Catalog API reference for constraints
on reques; CreateTableBucket body '[]' -> HTTP 400 BadRequestException "{\"message\":\"Invalid input. Consult
the S3 Tables and/or Iceberg REST Catalog API reference for constraints on reques; GET /buckets?maxBuckets=abc
-> HTTP 400 BadRequestException "{\"message\":\"Invalid input. Consult the S3 Tables and/or Iceberg REST
Catalog API reference for constraints on reques; GET /tables/<arn>?maxTables=abc -> HTTP 400
BadRequestException "{\"message\":\"Invalid input. Consult the S3 Tables and/or Iceberg REST Catalog API
reference for constraints on reques.
