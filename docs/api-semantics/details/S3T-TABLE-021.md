<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-021: Table storage class is a creation-time snapshot: inheriting table still INTELLIGENT_TIERING 66 polls/120 s after bucket flips to STANDARD
_Full entry and notes of one finding; its summary entry is in
[table-encryption-storage.md](../table-encryption-storage.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-021"></a>**S3T-TABLE-021** `immutable-field` · impact high · handled · verified 2026-10-15, re-verified
  **Table storage class is a creation-time snapshot: inheriting table still INTELLIGENT_TIERING 66 polls/120 s after bucket flips to STANDARD**
  With tables t1 (no config, read INTELLIGENT_TIERING at create), t3 (explicit IT), t4 (explicit STANDARD)
  under a bucket default of INTELLIGENT_TIERING, PutTableBucketStorageClass(STANDARD) -> HTTP 200 (bucket read
  showed STANDARD after 0.01 s). GetTableStorageClass polled: t1 every 1 s for 60 s then every 10 s to 120 s
  -> values ['INTELLIGENT_TIERING']; t3 -> ['INTELLIGENT_TIERING']; t4 -> ['STANDARD']. A table created after
  the flip without config (t2) reads STANDARD.
  - ACK: compare.is_ignored+delta_pre_compare, is_immutable, late_initialize · ops:
    PutTableBucketStorageClass, GetTableStorageClass, CreateTable · fields:
    storageClassConfiguration.storageClass
  - repro: Bucket IT -> CreateTable t1 (no config) -> PutTableBucketStorageClass(STANDARD) -> poll
    GetTableStorageClass(t1) for 120 s -> CreateTable t2 (no config)
  - measurements: poll_window_s=120, t1_polls=66, bucket_storage_class_visible_after_put_s=[0.01, 0.01], n=2
  - handling: handled via `generator.yaml:24-49; pkg/resource/table/hooks.go:69-75; generator.yaml:47-49; apis/v1alpha1/table.go:54-59`
  - related: [S3T-TABLEBUCKET-145](../tablebucket-encryption-storage.md#s3t-tablebucket-145), [S3T-TABLE-260](../table-encryption-storage.md#s3t-table-260), [S3T-TABLE-261](../table-encryption-storage.md#s3t-table-261), [S3T-TABLE-262](../table-encryption-storage.md#s3t-table-262), [S3T-TABLEBUCKET-084](../tablebucket-encryption-storage.md#s3t-tablebucket-084),
    [S3T-TABLE-020](../table-encryption-storage.md#s3t-table-020), [S3T-TABLE-057](../table-encryption-storage.md#s3t-table-057), [S3T-TABLE-058](../table-encryption-storage.md#s3t-table-058) · hypotheses: H-T-174, H-T-175 · evidence:
    table/sub-resources/storage-class-inheritance

## Notes

H-T-174 second half (snapshot >= 2 min) confirmed; H-T-175 (live tracking within 60 s) refuted. The bucket
default is read live by GetTableBucketStorageClass, the table read is not.

Full original title: Table storage class is a creation-time snapshot: after the bucket default flips
IT->STANDARD the inheriting table still reads INTELLIGENT_TIERING (66 polls/120 s)
