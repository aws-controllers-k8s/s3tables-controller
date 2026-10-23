<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-301: Explicit CreateTable AES256 overrides a bucket SSE-KMS default (200, reads AES256) even when the bucket key lacks the maintenance principal
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-301"></a>**S3T-TABLE-301** `server-default` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Explicit CreateTable AES256 overrides a bucket SSE-KMS default (200, reads AES256) even when the bucket key lacks the maintenance principal**
  Bucket default PutTableBucketEncryption(aws:kms K) -> HTTP 200 with K's policy lacking
  maintenance.s3tables.amazonaws.com. CreateTable(no encryptionConfiguration) -> BadRequestException (HTTP 400)
  'Insufficient access to perform table maintenance. Ensure that the key used to encrypt the table still
  exists, is active, and has a resource policy granting access to the S3 service principal
  'maintenance.s3tables.amazonaws.com'.' (table exists after reject: False);
  CreateTable(encryptionConfiguration AES256) -> HTTP 200, GetTableEncryption {"sseAlgorithm": "AES256"};
  CreateTable(explicit aws:kms K) -> BadRequestException (HTTP 400) 'Insufficient access to perform table
  maintenance. Ensure that the key used to encrypt the table still exists, is active, and has a resource
  policy granting access to the S3 service principal 'maintenance.s3tables.amazonaws.com'.'. After
  PutKeyPolicy granting maintenance.s3tables.amazonaws.com kms:GenerateDataKey/Decrypt (+10 s): omitted ->
  HTTP 200, GetTableEncryption {"sseAlgorithm": "aws:kms", "kmsKeyArn":
  "arn:aws:kms:us-west-2:<ACCOUNT>:key/4b033ad1-612c-46b0-99cd-cba251d5a1f5"}; explicit AES256 -> HTTP 200,
  reads {"sseAlgorithm": "AES256"}; explicit aws:kms K -> HTTP 200, reads {"sseAlgorithm": "aws:kms",
  "kmsKeyArn": "arn:aws:kms:us-west-2:<ACCOUNT>:key/4b033ad1-612c-46b0-99cd-cba251d5a1f5"}. Bucket default
  Put(AES256) -> HTTP 200; explicit aws:kms K on the AES256 bucket -> HTTP 200, reads {"sseAlgorithm":
  "aws:kms", "kmsKeyArn": "arn:aws:kms:us-west-2:<ACCOUNT>:key/4b033ad1-612c-46b0-99cd-cba251d5a1f5"}; omitted
  -> {"sseAlgorithm": "AES256"}; DeleteTableBucketEncryption -> HTTP 204.
  - ACK: late_initialize, compare.is_ignored+delta_pre_compare · ops: PutTableBucketEncryption, CreateTable,
    GetTableEncryption, PutKeyPolicy · fields: encryptionConfiguration.sseAlgorithm,
    encryptionConfiguration.kmsKeyArn
  - repro: CreateKey (default policy) -> PutTableBucketEncryption(aws:kms K) -> CreateTable with/without
    explicit AES256 -> PutKeyPolicy(+maintenance principal) -> repeat
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-057](../table-encryption-storage.md#s3t-table-057), [S3T-TABLE-239](../table-record-expiration.md#s3t-table-239), [S3T-TABLE-059](../table-encryption-storage.md#s3t-table-059), [S3T-TABLE-266](../table-encryption-storage.md#s3t-table-266), [S3T-TABLE-254](../table-maintenance.md#s3t-table-254), [S3T-TABLE-253](../table-maintenance.md#s3t-table-253),
    [S3T-TABLE-246](../table-maintenance.md#s3t-table-246), [S3T-TABLE-248](../table-encryption-storage.md#s3t-table-248), [S3T-TABLE-054](../table-maintenance.md#s3t-table-054), [S3T-TABLE-055](../table-maintenance.md#s3t-table-055), [S3T-TABLEBUCKET-119](../tablebucket-encryption-storage.md#s3t-tablebucket-119), [S3T-TABLE-235](../table-maintenance.md#s3t-table-235) · evidence:
    table/creative/kms-default-vs-explicit-aes256

## Notes

Cross-service analogy (DynamoDB TableClass / ElastiCache parameter defaults): a per-table explicit value equal
to the service default but different from the parent default. Mirrors the storage-class cell ([S3T-TABLE-266](../table-encryption-storage.md#s3t-table-266)).
Extends [S3T-TABLE-057](../table-encryption-storage.md#s3t-table-057) (inherit) and [S3T-TABLE-239](../table-record-expiration.md#s3t-table-239) (key grant missing -> 400) with the explicit-AES256 cells.
Fresh key, one bucket, n=1 per cell.
