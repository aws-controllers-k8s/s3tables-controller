<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-084: PutTableBucketStorageClass STANDARD->INTELLIGENT_TIERING->STANDARD (HTTP 200, visible on the first Get); existing tables keep their class
_Full entry and notes of one finding; its summary entry is in
[tablebucket-encryption-storage.md](../tablebucket-encryption-storage.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-084"></a>**S3T-TABLEBUCKET-084** `sub-resource-api` · impact high · handled · verified 2026-10-13
  **PutTableBucketStorageClass STANDARD->INTELLIGENT_TIERING->STANDARD (HTTP 200, visible on the first Get); existing tables keep their class**
  Table t1 created under the default -> GetTableStorageClass STANDARD. Put(INTELLIGENT_TIERING) -> HTTP 200;
  bucket Get at +0 s -> INTELLIGENT_TIERING, +1 s -> INTELLIGENT_TIERING. CreateTable t2 ->
  GetTableStorageClass INTELLIGENT_TIERING; t1 re-read -> STANDARD. Put(STANDARD) -> HTTP 200; bucket Get ->
  STANDARD; CreateTable t3 -> STANDARD; t2 re-read -> INTELLIGENT_TIERING.
  - ACK: custom_update, docs-only · ops: PutTableBucketStorageClass, GetTableBucketStorageClass, CreateTable,
    GetTableStorageClass · fields: storageClassConfiguration.storageClass
  - repro: CreateTable t1; Put(INTELLIGENT_TIERING); Get +0/+1 s; CreateTable t2; read t1/t2; Put(STANDARD);
    Get; CreateTable t3; read t3/t2
  - measurements: bucket_get_reflects_put_on_first_read=true, n=2
  - handling: handled via `generator.yaml:194-195; generator.yaml:284-285; pkg/resource/table_bucket/hooks.go:90-105; pkg/resource/table_bucket/hooks.go:508-537; pkg/resource/table_bucket/hooks.go:484-486; pkg/resource/table_bucket/hooks.go:519-522`
  - related: [S3T-TABLEBUCKET-009](../tablebucket-encryption-storage.md#s3t-tablebucket-009), [S3T-TABLEBUCKET-083](../tablebucket-encryption-storage.md#s3t-tablebucket-083), [S3T-TABLEBUCKET-085](../tablebucket-encryption-storage.md#s3t-tablebucket-085), [S3T-TABLE-161](../table-encryption-storage.md#s3t-table-161),
    [S3T-TABLEBUCKET-145](../tablebucket-encryption-storage.md#s3t-tablebucket-145), [S3T-TABLE-260](../table-encryption-storage.md#s3t-table-260), [S3T-TABLE-261](../table-encryption-storage.md#s3t-table-261), [S3T-TABLE-262](../table-encryption-storage.md#s3t-table-262), [S3T-TABLE-020](../table-encryption-storage.md#s3t-table-020), [S3T-TABLE-021](../table-encryption-storage.md#s3t-table-021),
    [S3T-TABLEBUCKET-146](../tablebucket-encryption-storage.md#s3t-tablebucket-146), [S3T-TABLEBUCKET-147](../tablebucket-encryption-storage.md#s3t-tablebucket-147), [S3T-TABLEBUCKET-117](../tablebucket-encryption-storage.md#s3t-tablebucket-117), [S3T-TABLEBUCKET-059](../tablebucket-encryption-storage.md#s3t-tablebucket-059), [S3T-TABLEBUCKET-008](../tablebucket-encryption-storage.md#s3t-tablebucket-008),
    [S3T-TABLE-266](../table-encryption-storage.md#s3t-table-266), [S3T-TABLEBUCKET-072](../tablebucket-encryption-storage.md#s3t-tablebucket-072), [S3T-TABLEBUCKET-073](../tablebucket-encryption-storage.md#s3t-tablebucket-073) · hypotheses: H-S-038, H-S-086 · evidence:
    tablebucket/sub-resources/storage-class-default-transitions

## Notes

H-S-038 confirmed. H-S-086 (contrarian: revert to STANDARD rejected) refuted: Put(STANDARD) after
INTELLIGENT_TIERING returned HTTP 200.

Full original title: PutTableBucketStorageClass STANDARD->INTELLIGENT_TIERING (HTTP 200) ->STANDARD (HTTP
200), visible on the first Get; tables keep their class (t1 STANDARD, t2 INTELLIGENT_TIERING)
