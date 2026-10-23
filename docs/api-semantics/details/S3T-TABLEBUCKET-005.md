<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-005: GetTableBucketEncryption on a bucket created without encryptionConfiguration -> HTTP 200 {"sseAlgorithm": "AES256"}, no kmsKeyArn member
_Full entry and notes of one finding; its summary entry is in
[tablebucket-encryption-storage.md](../tablebucket-encryption-storage.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-005"></a>**S3T-TABLEBUCKET-005** `server-default` · impact high · handled · verified 2026-10-13
  **GetTableBucketEncryption on a bucket created without encryptionConfiguration -> HTTP 200 {"sseAlgorithm": "AES256"}, no kmsKeyArn member**
  Bucket created with name+tags only: GetTableBucketEncryption at 0 s -> HTTP 200, raw JSON
  {"encryptionConfiguration": {"sseAlgorithm": "AES256"}}; at +3 s -> HTTP 200, raw JSON
  {"encryptionConfiguration": {"sseAlgorithm": "AES256"}}. After PutTableBucketEncryption(AES256) (-> HTTP 200)
  the read returned {"encryptionConfiguration": {"sseAlgorithm": "AES256"}}; DeleteTableBucketEncryption ->
  HTTP 204 (raw status 204, body ''); GetTableBucketEncryption right after the delete -> HTTP 200
  {"encryptionConfiguration": {"sseAlgorithm": "AES256"}}, at +2 s -> HTTP 200 {"encryptionConfiguration":
  {"sseAlgorithm": "AES256"}}; a second DeleteTableBucketEncryption -> NotFoundException (HTTP 404).
  - ACK: late_initialize, compare.is_ignored+delta_pre_compare, exceptions.404 · ops:
    GetTableBucketEncryption, PutTableBucketEncryption, DeleteTableBucketEncryption · fields:
    encryptionConfiguration.sseAlgorithm, encryptionConfiguration.kmsKeyArn
  - repro: CreateTableBucket(name) -> GetTableBucketEncryption; PutTableBucketEncryption(AES256) -> Get;
    DeleteTableBucketEncryption -> Get (0 s, +2 s)
  - handling: handled via `pkg/resource/table_bucket/hooks.go:69-88; pkg/resource/table_bucket/hooks.go:473-506; generator.yaml:213-228; pkg/resource/table_bucket/manager.go:53`
  - related: [S3T-TABLEBUCKET-048](../tablebucket-encryption-storage.md#s3t-tablebucket-048), [S3T-TABLEBUCKET-037](../tablebucket-encryption-storage.md#s3t-tablebucket-037), [S3T-TABLEBUCKET-049](../tablebucket-encryption-storage.md#s3t-tablebucket-049), [S3T-TABLEBUCKET-051](../tablebucket-encryption-storage.md#s3t-tablebucket-051),
    [S3T-TABLEBUCKET-052](../tablebucket-encryption-storage.md#s3t-tablebucket-052), [S3T-TABLEBUCKET-072](../tablebucket-encryption-storage.md#s3t-tablebucket-072), [S3T-TABLE-010](../table-encryption-storage.md#s3t-table-010), [S3T-TABLE-011](../table-encryption-storage.md#s3t-table-011), [S3T-TABLEBUCKET-083](../tablebucket-encryption-storage.md#s3t-tablebucket-083) · hypotheses:
    H-B-006, H-B-007 · evidence: tablebucket/round-trip/create-identity-defaults

## Notes

Confirms H-B-006 (default materialized, kmsKeyArn member absent) and refutes H-B-007 (no NotFoundException for
the unset default).

Full original title: GetTableBucketEncryption on a bucket created without encryptionConfiguration returns HTTP
200 with the materialized default {"sseAlgorithm": "AES256"}
