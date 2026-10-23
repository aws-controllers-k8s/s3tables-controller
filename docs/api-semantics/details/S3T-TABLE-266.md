<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-266: Explicit CreateTable storageClass=STANDARD wins over the bucket default INTELLIGENT_TIERING (reads STANDARD); omitted inherits the default
_Full entry and notes of one finding; its summary entry is in
[table-encryption-storage.md](../table-encryption-storage.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-266"></a>**S3T-TABLE-266** `server-default` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Explicit CreateTable storageClass=STANDARD wins over the bucket default INTELLIGENT_TIERING (reads STANDARD); omitted inherits the default**
  PutTableBucketStorageClass(INTELLIGENT_TIERING) -> HTTP 200.
  CreateTable(storageClassConfiguration={storageClass: STANDARD}) -> 200; GetTableStorageClass ->
  {"storageClass": "STANDARD"}. CreateTable without storageClassConfiguration -> GetTableStorageClass ->
  {"storageClass": "INTELLIGENT_TIERING"}. After PutTableBucketStorageClass(STANDARD): CreateTable(explicit
  INTELLIGENT_TIERING) -> {"storageClass": "INTELLIGENT_TIERING"}.
  - ACK: late_initialize, compare.is_ignored+delta_pre_compare · ops: PutTableBucketStorageClass, CreateTable,
    GetTableStorageClass · fields: storageClassConfiguration.storageClass
  - repro: PutTableBucketStorageClass(INTELLIGENT_TIERING) -> CreateTable(storageClassConfiguration STANDARD)
    -> GetTableStorageClass
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-020](../table-encryption-storage.md#s3t-table-020), [S3T-TABLE-059](../table-encryption-storage.md#s3t-table-059), [S3T-TABLE-161](../table-encryption-storage.md#s3t-table-161), [S3T-TABLEBUCKET-084](../tablebucket-encryption-storage.md#s3t-tablebucket-084), [S3T-TABLEBUCKET-072](../tablebucket-encryption-storage.md#s3t-tablebucket-072),
    [S3T-TABLEBUCKET-073](../tablebucket-encryption-storage.md#s3t-tablebucket-073) · evidence: table/creative/bucket-maintenance-propagation

## Notes

Cross-service analogy (DynamoDB TableClass default / ElastiCache parameter defaults): whether an explicit
per-table value is honored when it equals the service default but differs from the parent default. Extends
[S3T-TABLE-020](../table-encryption-storage.md#s3t-table-020) and [S3T-TABLE-059](../table-encryption-storage.md#s3t-table-059) (explicit values on a bucket WITHOUT a conflicting default).

Full original title: Bucket default INTELLIGENT_TIERING + CreateTable(explicit STANDARD) -> table reads
STANDARD (explicit overrides bucket default); omitted -> INTELLIGENT_TIERING
