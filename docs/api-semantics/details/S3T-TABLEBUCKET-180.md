<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-180: SSE-KMS bucket with a default-policy key: CreateTableBucket 200 but CreateTable 400 until the key policy grants the maintenance principal
_Full entry and notes of one finding; its summary entry is in
[tablebucket-encryption-storage.md](../tablebucket-encryption-storage.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-180"></a>**S3T-TABLEBUCKET-180** `prerequisite` · impact high · handled · verified 2026-10-14
  **SSE-KMS bucket with a default-policy key: CreateTableBucket 200 but CreateTable 400 until the key policy grants the maintenance principal**
  KMS CreateKey without a Policy (default key policy: only the account root statement) then
  CreateTableBucket(encryptionConfiguration={aws:kms, that key}) -> HTTP 200; GetTableBucketEncryption -> HTTP 200.
  In that bucket CreateTable(format=ICEBERG, no metadata) -> BadRequestException (HTTP 400) 'Insufficient
  access to perform table maintenance. Ensure that the key used to encrypt the table still exists, is active,
  and has a resource policy granting access to the S3 service principal
  'maintenance.s3tables.amazonaws.com'.'; CreateTable(metadata.iceberg.schema 2 fields) -> BadRequestException
  (HTTP 400) 'Insufficient access to perform table maintenance. Ensure that the key used to encrypt the table
  still exists, is active, and has a resource policy granting access to the S3 service principal
  'maintenance.s3tables.amazonaws.com'.'. After PutKeyPolicy adding Principal Service
  maintenance.s3tables.amazonaws.com with kms:GenerateDataKey+kms:Decrypt (Condition aws:SourceAccount) (+5
  s): CreateTable(no metadata) -> HTTP 200; CreateTable(schema) -> HTTP 200.
  - ACK: terminal_codes, docs-only, references · ops: CreateTableBucket, CreateTable, GetTableBucketEncryption
    · fields: encryptionConfiguration.sseAlgorithm, encryptionConfiguration.kmsKeyArn, metadata.iceberg.schema
  - repro: kms CreateKey (no Policy) -> CreateTableBucket(aws:kms, key) -> CreateNamespace -> CreateTable x2
    -> PutKeyPolicy(+maintenance principal) -> CreateTable x2
  - measurements: n_create_table_before=2, n_create_table_after=2, n=2
  - handling: handled via `generator.yaml:220-225; test/e2e/tests/test_table_bucket.py:158-162`
  - related: [S3T-TABLEBUCKET-036](../tablebucket-encryption-storage.md#s3t-tablebucket-036), [S3T-TABLEBUCKET-119](../tablebucket-encryption-storage.md#s3t-tablebucket-119), [S3T-TABLEBUCKET-120](../tablebucket-encryption-storage.md#s3t-tablebucket-120), [S3T-TABLEBUCKET-182](../tablebucket-encryption-storage.md#s3t-tablebucket-182),
    [S3T-TABLEBUCKET-183](../tablebucket-encryption-storage.md#s3t-tablebucket-183) · evidence: tablebucket/sub-resources/fu-01-kms-key-policy-disable

## Notes

FU-01 part 1. The caller is an account admin, so caller-side KMS permissions are not the limiting factor here;
the only difference between the two halves is the maintenance.s3tables.amazonaws.com statement in the key
policy. The grant changed the CreateTable outcome. Both CreateTable variants (with and without metadata)
failed identically, so the check is done at table creation regardless of whether a metadata.json is written;
the message names the maintenance principal explicitly. Controller implication: a bucket with aws:kms needs
the key policy prerequisite before any Table child can be created; the error is terminal until the key policy
changes (fixed within 5 s here).
