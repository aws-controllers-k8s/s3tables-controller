<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-052: DeleteTableBucketEncryption on a bucket created WITH encryptionConfiguration (aws:kms) returns HTTP 204...
_Full entry and notes of one finding; its summary entry is in
[tablebucket-encryption-storage.md](../tablebucket-encryption-storage.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-052"></a>**S3T-TABLEBUCKET-052** `delete-semantics` · impact medium · handled · verified 2026-10-13
  **DeleteTableBucketEncryption on a bucket created WITH encryptionConfiguration (aws:kms) returns HTTP 204...**
  Bucket created with encryptionConfiguration aws:kms: Get -> '200
  {"encryptionConfiguration":{"kmsKeyArn":"<kms-key-arn>","sseAlgorithm":"aws:kms"}}'; Delete -> HTTP 204; Get
  -> '200 {"encryptionConfiguration":{"sseAlgorithm":"AES256"}}' (equals the never-configured body). Bucket
  created with explicit AES256: Get -> '200 {"encryptionConfiguration":{"sseAlgorithm":"AES256"}}' (identical
  to the never-configured body); Delete -> HTTP 204; Get -> '200
  {"encryptionConfiguration":{"sseAlgorithm":"AES256"}}'.
  - ACK: is_immutable, custom_delete · ops: CreateTableBucket, DeleteTableBucketEncryption,
    GetTableBucketEncryption · fields: encryptionConfiguration
  - repro: CreateTableBucket(encryptionConfiguration=aws:kms) -> Get -> Delete -> Get;
    CreateTableBucket(AES256) -> Get -> Delete -> Get
  - measurements: n=1
  - handling: handled via `pkg/resource/table_bucket/hooks.go:69-88; pkg/resource/table_bucket/hooks.go:473-506; generator.yaml:213-228; pkg/resource/table_bucket/manager.go:53; pkg/resource/table_bucket/hooks.go:484-486; pkg/resource/table_bucket/hooks.go:519-522`
  - related: [S3T-TABLEBUCKET-005](../tablebucket-encryption-storage.md#s3t-tablebucket-005), [S3T-TABLEBUCKET-037](../tablebucket-encryption-storage.md#s3t-tablebucket-037), [S3T-TABLEBUCKET-049](../tablebucket-encryption-storage.md#s3t-tablebucket-049), [S3T-TABLEBUCKET-051](../tablebucket-encryption-storage.md#s3t-tablebucket-051),
    [S3T-TABLEBUCKET-072](../tablebucket-encryption-storage.md#s3t-tablebucket-072), [S3T-TABLEBUCKET-050](../tablebucket-encryption-storage.md#s3t-tablebucket-050), [S3T-TABLEBUCKET-083](../tablebucket-encryption-storage.md#s3t-tablebucket-083), [S3T-TABLEBUCKET-103](../tablebucket-encryption-storage.md#s3t-tablebucket-103), [S3T-TABLE-059](../table-encryption-storage.md#s3t-table-059) ·
    hypotheses: H-S-132 · evidence: tablebucket/sub-resources/encryption-get-put-delete

## Notes

H-S-132 (contrarian: create-time encryption stickier) refuted: create-time encryption is removable by Delete.
Explicit AES256 at create is indistinguishable from the default in Get.

Full original title: DeleteTableBucketEncryption on a bucket created WITH encryptionConfiguration (aws:kms)
returns HTTP 204, same as post-create Put then Delete (HTTP 204)
