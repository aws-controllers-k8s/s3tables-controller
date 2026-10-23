<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-085: GetTableBucketStorageClass returns the BUCKET default (STANDARD) even with an INTELLIGENT_TIERING table; table ARN as tableBucketARN -> 400
_Full entry and notes of one finding; its summary entry is in
[tablebucket-encryption-storage.md](../tablebucket-encryption-storage.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-085"></a>**S3T-TABLEBUCKET-085** `shape-mismatch` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **GetTableBucketStorageClass returns the BUCKET default (STANDARD) even with an INTELLIGENT_TIERING table; table ARN as tableBucketARN -> 400**
  CreateTable with storageClassConfiguration INTELLIGENT_TIERING in a STANDARD-default bucket -> HTTP 200,
  GetTableStorageClass -> INTELLIGENT_TIERING. GetTableBucketStorageClass(bucket ARN) -> STANDARD.
  GetTableBucketStorageClass(tableBucketARN=<table ARN>) -> BadRequestException (HTTP 400) 'The specified
  bucket name isn't valid. Specify a different bucket name, and then try again.'.
  - ACK: docs-only, is_arn_primary_key · ops: GetTableBucketStorageClass, CreateTable, GetTableStorageClass ·
    fields: tableBucketARN, storageClassConfiguration.storageClass
  - repro: CreateTable(storageClass=INTELLIGENT_TIERING) in a STANDARD bucket;
    GetTableBucketStorageClass(bucket ARN); GetTableBucketStorageClass(table ARN)
  - measurements: n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-009](../tablebucket-encryption-storage.md#s3t-tablebucket-009), [S3T-TABLEBUCKET-084](../tablebucket-encryption-storage.md#s3t-tablebucket-084), [S3T-TABLEBUCKET-083](../tablebucket-encryption-storage.md#s3t-tablebucket-083), [S3T-TABLE-161](../table-encryption-storage.md#s3t-table-161), [S3T-TABLE-062](../table-encryption-storage.md#s3t-table-062) ·
    hypotheses: H-S-039 · evidence: tablebucket/sub-resources/storage-class-default-transitions

## Notes

H-S-039 confirmed: the model doc ('storage class configuration for a specific table') is a copy-paste; the op
returns the bucket default only.

Full original title: GetTableBucketStorageClass returns the BUCKET default (STANDARD) with an
INTELLIGENT_TIERING table present; a table ARN as tableBucketARN -> BadRequestException (HTTP 400)
