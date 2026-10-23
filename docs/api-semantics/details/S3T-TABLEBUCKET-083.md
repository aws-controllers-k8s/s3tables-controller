<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-083: GetTableBucketStorageClass on a fresh bucket -> 200 {"storageClass": "STANDARD"}; after an explicit Put(STANDARD) the body is byte-identical
_Full entry and notes of one finding; its summary entry is in
[tablebucket-encryption-storage.md](../tablebucket-encryption-storage.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-083"></a>**S3T-TABLEBUCKET-083** `server-default` · impact high · handled · verified 2026-10-13
  **GetTableBucketStorageClass on a fresh bucket -> 200 {"storageClass": "STANDARD"}; after an explicit Put(STANDARD) the body is byte-identical**
  CreateTableBucket(name, tags only) -> GetTableBucketStorageClass at +0 s: '200
  {"storageClassConfiguration":{"storageClass":"STANDARD"}}' and at +5 s: '200
  {"storageClassConfiguration":{"storageClass":"STANDARD"}}' (identical). PutTableBucketStorageClass(STANDARD)
  -> HTTP 200; Get -> '200 {"storageClassConfiguration":{"storageClass":"STANDARD"}}' (identical to the fresh
  body: no isDefault/source marker).
  - ACK: late_initialize, compare.nil_equals_zero_value · ops: GetTableBucketStorageClass,
    PutTableBucketStorageClass, CreateTableBucket · fields: storageClassConfiguration.storageClass
  - repro: CreateTableBucket -> GetTableBucketStorageClass (+0 s, +5 s, raw body) -> Put(STANDARD) -> Get (raw
    body)
  - measurements: n=2
  - handling: handled via `pkg/resource/table_bucket/hooks.go:90-105; pkg/resource/table_bucket/hooks.go:508-537; generator.yaml:213-228; pkg/resource/table_bucket/manager.go:53; pkg/resource/table_bucket/hooks.go:484-486; pkg/resource/table_bucket/hooks.go:519-522`
  - related: [S3T-TABLEBUCKET-006](../tablebucket-encryption-storage.md#s3t-tablebucket-006), [S3T-TABLEBUCKET-050](../tablebucket-encryption-storage.md#s3t-tablebucket-050), [S3T-TABLEBUCKET-052](../tablebucket-encryption-storage.md#s3t-tablebucket-052), [S3T-TABLEBUCKET-103](../tablebucket-encryption-storage.md#s3t-tablebucket-103),
    [S3T-TABLE-059](../table-encryption-storage.md#s3t-table-059), [S3T-TABLEBUCKET-009](../tablebucket-encryption-storage.md#s3t-tablebucket-009), [S3T-TABLEBUCKET-084](../tablebucket-encryption-storage.md#s3t-tablebucket-084), [S3T-TABLEBUCKET-085](../tablebucket-encryption-storage.md#s3t-tablebucket-085), [S3T-TABLE-161](../table-encryption-storage.md#s3t-table-161),
    [S3T-TABLE-010](../table-encryption-storage.md#s3t-table-010), [S3T-TABLE-011](../table-encryption-storage.md#s3t-table-011), [S3T-TABLEBUCKET-005](../tablebucket-encryption-storage.md#s3t-tablebucket-005) · hypotheses: H-S-036, H-S-037, H-S-137 · evidence:
    tablebucket/sub-resources/storage-class-default-transitions

## Notes

H-S-036 (200 STANDARD materialized) confirmed; H-S-037 (contrarian 404 until first Put) refuted; H-S-137
(explicit STANDARD indistinguishable from implicit) confirmed.

Full original title: GetTableBucketStorageClass on a fresh bucket returns 200
{'storageClassConfiguration':{'storageClass':'STANDARD'}}; after an explicit Put(STANDARD) the body is
byte-identical
