<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-020: Bucket default INTELLIGENT_TIERING: CreateTable without storageClassConfiguration inherits it; explicit STANDARD reads STANDARD
_Full entry and notes of one finding; its summary entry is in
[table-encryption-storage.md](../table-encryption-storage.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-020"></a>**S3T-TABLE-020** `server-default` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Bucket default INTELLIGENT_TIERING: CreateTable without storageClassConfiguration inherits it; explicit STANDARD reads STANDARD**
  PutTableBucketStorageClass(INTELLIGENT_TIERING) -> HTTP 200 (GetTableBucketStorageClass showed it after 0.01
  s; fresh bucket read before the put: HTTP 200 {"storageClassConfiguration": {"storageClass": "STANDARD"}};
  put with storageClass=GLACIER -> ValidationException (HTTP 400) '1 validation error detected: Value at
  'storageClassConfiguration.storageClass' failed to satisfy constraint: Member must satisfy enum value set:
  [INTELLIGENT_TIERING, STANDARD]'). Then CreateTable(no storageClassConfiguration) -> HTTP 200,
  GetTableStorageClass -> INTELLIGENT_TIERING; CreateTable(explicit INTELLIGENT_TIERING) -> HTTP 200, read
  INTELLIGENT_TIERING; CreateTable(explicit STANDARD) -> HTTP 200, read STANDARD.
  - ACK: late_initialize, compare.is_ignored+delta_pre_compare · ops: PutTableBucketStorageClass,
    GetTableBucketStorageClass, CreateTable, GetTableStorageClass · fields:
    storageClassConfiguration.storageClass
  - repro: PutTableBucketStorageClass(IT) -> CreateTable(no config), CreateTable(explicit IT),
    CreateTable(explicit STANDARD) -> GetTableStorageClass on each
  - measurements: bucket_storage_class_visible_after_put_s=[0.01], n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-087](../tablebucket-encryption-storage.md#s3t-tablebucket-087), [S3T-TABLEBUCKET-024](../tablebucket-encryption-storage.md#s3t-tablebucket-024), [S3T-TABLEBUCKET-023](../tablebucket-encryption-storage.md#s3t-tablebucket-023), [S3T-TABLE-012](../table-encryption-storage.md#s3t-table-012), [S3T-TABLE-013](../table-encryption-storage.md#s3t-table-013),
    [S3T-TABLEBUCKET-145](../tablebucket-encryption-storage.md#s3t-tablebucket-145), [S3T-TABLE-260](../table-encryption-storage.md#s3t-table-260), [S3T-TABLE-261](../table-encryption-storage.md#s3t-table-261), [S3T-TABLE-262](../table-encryption-storage.md#s3t-table-262), [S3T-TABLEBUCKET-084](../tablebucket-encryption-storage.md#s3t-tablebucket-084), [S3T-TABLE-021](../table-encryption-storage.md#s3t-table-021),
    [S3T-TABLE-266](../table-encryption-storage.md#s3t-table-266), [S3T-TABLE-161](../table-encryption-storage.md#s3t-table-161), [S3T-TABLEBUCKET-072](../tablebucket-encryption-storage.md#s3t-tablebucket-072), [S3T-TABLEBUCKET-073](../tablebucket-encryption-storage.md#s3t-tablebucket-073) · hypotheses: H-T-174, H-T-177 ·
    evidence: table/sub-resources/storage-class-inheritance

## Notes

H-T-174 first half (inheritance at create) confirmed; H-T-177 (explicit STANDARD under an IT bucket default)
confirmed. No DeleteTableBucketStorageClass operation exists in the API, so 'deleting the bucket config'
cannot be tested.

Full original title: Bucket default INTELLIGENT_TIERING: CreateTable without storageClassConfiguration reads
INTELLIGENT_TIERING; explicit STANDARD reads STANDARD
