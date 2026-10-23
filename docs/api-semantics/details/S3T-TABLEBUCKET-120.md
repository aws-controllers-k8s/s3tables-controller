<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-120: Bucket SSE-KMS applies only to tables created after the Put (t0 before: AES256; t1 after: aws:kms)...
_Full entry and notes of one finding; its summary entry is in
[tablebucket-encryption-storage.md](../tablebucket-encryption-storage.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-120"></a>**S3T-TABLEBUCKET-120** `sub-resource-api` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Bucket SSE-KMS applies only to tables created after the Put (t0 before: AES256; t1 after: aws:kms)...**
  After granting maintenance.s3tables.amazonaws.com kms:GenerateDataKey+kms:Decrypt on the key (30 s wait),
  CreateTable t1 -> HTTP 200 (attempts: [(0.6, 'HTTP 200')]). GetTableEncryption: t0 (created before
  PutTableBucketEncryption) -> {'sseAlgorithm': 'AES256'}; t1 (created after) -> {'sseAlgorithm': 'aws:kms',
  'kmsKeyArn': 'arn:aws:kms:us-west-2:<ACCOUNT>:key/a6d732a8-1f64-4a80-b809-5091ee403efb'}.
  DeleteTableBucketEncryption -> HTTP 204, GetTableBucketEncryption -> {'sseAlgorithm': 'AES256'}; CreateTable
  t2 -> HTTP 200, its encryption -> {'sseAlgorithm': 'AES256'}; t1 afterwards -> {'sseAlgorithm': 'aws:kms',
  'kmsKeyArn': 'arn:aws:kms:us-west-2:<ACCOUNT>:key/a6d732a8-1f64-4a80-b809-5091ee403efb'}.
  - ACK: e2e-timing, docs-only · ops: PutTableBucketEncryption, DeleteTableBucketEncryption, CreateTable,
    GetTableEncryption, GetTableBucketEncryption · fields: encryptionConfiguration.sseAlgorithm,
    encryptionConfiguration.kmsKeyArn
  - repro: CreateTable t0 -> PutTableBucketEncryption(aws:kms) -> grant maintenance principal -> CreateTable
    t1 -> GetTableEncryption t0,t1 -> DeleteTableBucketEncryption -> CreateTable t2 -> GetTableEncryption
    t1,t2
  - measurements: key_policy_wait_s=30, create_table_attempts_after_grant=1, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-119](../tablebucket-encryption-storage.md#s3t-tablebucket-119), [S3T-TABLEBUCKET-180](../tablebucket-encryption-storage.md#s3t-tablebucket-180), [S3T-TABLE-255](../table-encryption-storage.md#s3t-table-255), [S3T-TABLE-259](../table-encryption-storage.md#s3t-table-259), [S3T-TABLE-257](../table-encryption-storage.md#s3t-table-257),
    [S3T-TABLE-057](../table-encryption-storage.md#s3t-table-057), [S3T-TABLE-010](../table-encryption-storage.md#s3t-table-010) · hypotheses: H-S-009, H-S-007 · evidence:
    tablebucket/dependencies/kms-key-policy-gate

## Notes

H-S-009 confirmed. H-S-007 second half (CreateTable succeeds once the key policy grants the maintenance
principal; the bucket kept the KMS configuration): confirmed.

Full original title: Bucket SSE-KMS applies only to tables created after the Put (t0 before: AES256; t1 after:
aws:kms); after DeleteTableBucketEncryption new table t2: AES256, t1 keeps aws:kms
