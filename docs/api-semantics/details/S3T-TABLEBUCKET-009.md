<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-009: storageClassConfiguration is mutable via PutTableBucketStorageClass (HTTP 200, visible after 0.01 s); no Delete op exists to unset it
_Full entry and notes of one finding; its summary entry is in
[tablebucket-encryption-storage.md](../tablebucket-encryption-storage.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-009"></a>**S3T-TABLEBUCKET-009** `sub-resource-api` · impact medium · handled · verified 2026-10-13
  **storageClassConfiguration is mutable via PutTableBucketStorageClass (HTTP 200, visible after 0.01 s); no Delete op exists to unset it**
  Bucket created with storageClass=INTELLIGENT_TIERING read back {"storageClass": "INTELLIGENT_TIERING"}
  immediately. PutTableBucketStorageClass(STANDARD) -> HTTP 200 (raw body ''); GetTableBucketStorageClass
  showed STANDARD after 0.01 s (1 s polling, first poll immediately). Re-sending STANDARD unchanged -> HTTP 200.
  PutTableBucketStorageClass(INTELLIGENT_TIERING) -> HTTP 200, visible after 0.01 s. The botocore model
  (apiVersion 2018-05-10) has no DeleteTableBucketStorageClass (present=False) and no Update*/Rename* bucket
  operation ([]).
  - ACK: custom_update, is_immutable · ops: CreateTableBucket, PutTableBucketStorageClass,
    GetTableBucketStorageClass · fields: storageClassConfiguration.storageClass, name
  - repro: CreateTableBucket(storageClass=INTELLIGENT_TIERING) -> PutTableBucketStorageClass(STANDARD) -> poll
    Get -> Put(INTELLIGENT_TIERING) again
  - measurements: put_to_visible_s=[0.01, 0.01], n=2
  - handling: handled via `generator.yaml:194-195; generator.yaml:284-285; pkg/resource/table_bucket/hooks.go:90-105; pkg/resource/table_bucket/hooks.go:508-537; pkg/resource/table_bucket/hooks.go:484-486; pkg/resource/table_bucket/hooks.go:519-522`
  - related: [S3T-TABLEBUCKET-084](../tablebucket-encryption-storage.md#s3t-tablebucket-084), [S3T-TABLEBUCKET-083](../tablebucket-encryption-storage.md#s3t-tablebucket-083), [S3T-TABLEBUCKET-085](../tablebucket-encryption-storage.md#s3t-tablebucket-085), [S3T-TABLE-161](../table-encryption-storage.md#s3t-table-161),
    [S3T-TABLEBUCKET-123](../tablebucket-encryption-storage.md#s3t-tablebucket-123), [S3T-TABLEBUCKET-073](../tablebucket-encryption-storage.md#s3t-tablebucket-073), [S3T-TABLEBUCKET-008](../tablebucket-encryption-storage.md#s3t-tablebucket-008), [S3T-TABLEBUCKET-037](../tablebucket-encryption-storage.md#s3t-tablebucket-037), [S3T-TABLE-061](../table-encryption-storage.md#s3t-table-061),
    [S3T-TABLEBUCKET-146](../tablebucket-encryption-storage.md#s3t-tablebucket-146), [S3T-TABLEBUCKET-147](../tablebucket-encryption-storage.md#s3t-tablebucket-147), [S3T-TABLEBUCKET-117](../tablebucket-encryption-storage.md#s3t-tablebucket-117), [S3T-TABLEBUCKET-059](../tablebucket-encryption-storage.md#s3t-tablebucket-059) · hypotheses: H-B-009,
    H-B-029, H-B-071 · evidence: tablebucket/round-trip/create-identity-defaults

## Notes

Confirms H-B-009 (not create-only; both directions accepted; the way back to the default is Put STANDARD).
Confirms the static H-B-029 (no UpdateTableBucket/rename in the model: name, type, ownerAccountId,
tableBucketId have no mutation path). Partially covers H-B-071: PutTableBucketStorageClass re-sent unchanged
is accepted (see outcome above).

Full original title: storageClassConfiguration is mutable after create via PutTableBucketStorageClass (HTTP
200, visible after 0.01 s); no Delete operation exists to unset it
