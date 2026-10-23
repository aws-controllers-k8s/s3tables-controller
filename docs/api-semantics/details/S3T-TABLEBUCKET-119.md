<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-119: Key lacking the maintenance principal: PutTableBucketEncryption 200, CreateTableBucket(aws:kms) 200, then CreateTable BadRequestException
_Full entry and notes of one finding; its summary entry is in
[tablebucket-encryption-storage.md](../tablebucket-encryption-storage.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-119"></a>**S3T-TABLEBUCKET-119** `prerequisite` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Key lacking the maintenance principal: PutTableBucketEncryption 200, CreateTableBucket(aws:kms) 200, then CreateTable BadRequestException**
  Customer key whose policy grants only the account root (no maintenance.s3tables.amazonaws.com statement).
  PutTableBucketEncryption(aws:kms, key) -> HTTP 200; GetTableBucketEncryption -> {'sseAlgorithm': 'aws:kms',
  'kmsKeyArn': 'arn:aws:kms:us-west-2:<ACCOUNT>:key/a6d732a8-1f64-4a80-b809-5091ee403efb'}. CreateTableBucket
  with the same encryptionConfiguration -> HTTP 200; its GetTableBucketEncryption -> {'sseAlgorithm':
  'aws:kms', 'kmsKeyArn': 'arn:aws:kms:us-west-2:<ACCOUNT>:key/a6d732a8-1f64-4a80-b809-5091ee403efb'}
  (byte-identical to the post-create Put body). CreateTable(format=ICEBERG) in the Put bucket ->
  BadRequestException (HTTP 400) 'Insufficient access to perform table maintenance. Ensure that the key used
  to encrypt the table still exists, is active, and has a resource policy granting access to the S3 service
  principal 'maintenance.s3tables.amazona'.
  - ACK: terminal_codes, docs-only · ops: PutTableBucketEncryption, CreateTableBucket,
    GetTableBucketEncryption, CreateTable · fields: encryptionConfiguration.kmsKeyArn,
    encryptionConfiguration.sseAlgorithm
  - repro: CreateKey(policy: root only) -> PutTableBucketEncryption(aws:kms) -> Get ->
    CreateTableBucket(encryptionConfiguration aws:kms) -> Get -> CreateTable
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-120](../tablebucket-encryption-storage.md#s3t-tablebucket-120), [S3T-TABLEBUCKET-180](../tablebucket-encryption-storage.md#s3t-tablebucket-180), [S3T-TABLEBUCKET-139](../tablebucket-encryption-storage.md#s3t-tablebucket-139), [S3T-TABLEBUCKET-142](../tablebucket-encryption-storage.md#s3t-tablebucket-142),
    [S3T-TABLE-256](../table-maintenance.md#s3t-table-256), [S3T-TABLE-257](../table-encryption-storage.md#s3t-table-257), [S3T-TABLEBUCKET-134](../tablebucket-encryption-storage.md#s3t-tablebucket-134), [S3T-TABLE-055](../table-maintenance.md#s3t-table-055), [S3T-TABLE-054](../table-maintenance.md#s3t-table-054), [S3T-TABLE-248](../table-encryption-storage.md#s3t-table-248),
    [S3T-TABLE-246](../table-maintenance.md#s3t-table-246), [S3T-TABLE-301](../table-maintenance.md#s3t-table-301), [S3T-TABLE-254](../table-maintenance.md#s3t-table-254) · hypotheses: H-S-007, H-S-083, H-S-012 · evidence:
    tablebucket/dependencies/kms-key-policy-gate

## Notes

H-S-083 (the Put itself validates the key policy): refuted. H-S-007 (Put succeeds, only CreateTable fails with
403 naming the maintenance principal/KMS): qualified: Put HTTP 200, CreateTable BadRequestException/400.
H-S-012 (create-time and Put bodies identical; create-time KMS unvalidated): confirmed.

Full original title: SSE-KMS with a key lacking the maintenance principal: PutTableBucketEncryption -> HTTP
200, CreateTableBucket(encryptionConfiguration) -> HTTP 200, CreateTable in the bucket ->
BadRequestException/400
