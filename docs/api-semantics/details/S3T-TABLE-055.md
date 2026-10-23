<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-055: Bucket default SSE-KMS with a key lacking the maintenance principal: PutTableBucketEncryption 200, CreateTable(no encryption config) 400
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-055"></a>**S3T-TABLE-055** `prerequisite` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Bucket default SSE-KMS with a key lacking the maintenance principal: PutTableBucketEncryption 200, CreateTable(no encryption config) 400**
  PutTableBucketEncryption(aws:kms, key with default policy only) -> HTTP 200; GetTableBucketEncryption
  afterwards -> HTTP 200 {"encryptionConfiguration": {"sseAlgorithm": "aws:kms", "kmsKeyArn":
  "arn:aws:kms:us-west-2:<ACCOUNT>:key/44faa909-a32b-42d8-baed-eec5099b107a"}}. CreateTable(ICEBERG, no
  encryptionConfiguration) under that bucket -> BadRequestException (HTTP 400) 'Insufficient access to perform
  table maintenance. Ensure that the key used to encrypt the table still exists, is active, and has a resource
  policy granting access to the S3 service principal 'maintenance.s3tables.amazonaws.com'.' (header
  'BadRequestException:http://internal.amazon.com/coral/com.amazon.s3tablebuckets/'). GetTable(triple) ->
  NotFoundException (HTTP 404) 'The specified table does not exist.'; GetTableEncryption -> NotFoundException
  (HTTP 404) 'The specified table does not exist.' .
  - ACK: terminal_codes, references · ops: PutTableBucketEncryption, GetTableBucketEncryption, CreateTable,
    GetTable · fields: encryptionConfiguration.kmsKeyArn
  - repro: PutTableBucketEncryption(aws:kms, key without maintenance statement) -> CreateTable without
    encryptionConfiguration -> GetTable
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-139](../tablebucket-encryption-storage.md#s3t-tablebucket-139), [S3T-TABLEBUCKET-142](../tablebucket-encryption-storage.md#s3t-tablebucket-142), [S3T-TABLE-256](../table-maintenance.md#s3t-table-256), [S3T-TABLE-257](../table-encryption-storage.md#s3t-table-257), [S3T-TABLEBUCKET-134](../tablebucket-encryption-storage.md#s3t-tablebucket-134),
    [S3T-TABLEBUCKET-119](../tablebucket-encryption-storage.md#s3t-tablebucket-119), [S3T-TABLE-054](../table-maintenance.md#s3t-table-054), [S3T-TABLE-056](../table-maintenance.md#s3t-table-056), [S3T-TABLE-057](../table-encryption-storage.md#s3t-table-057), [S3T-TABLE-141](../table-policy.md#s3t-table-141), [S3T-TABLE-248](../table-encryption-storage.md#s3t-table-248),
    [S3T-TABLE-246](../table-maintenance.md#s3t-table-246), [S3T-TABLE-301](../table-maintenance.md#s3t-table-301), [S3T-TABLE-254](../table-maintenance.md#s3t-table-254) · hypotheses: H-T-155 · evidence:
    table/dependencies/kms-maintenance-principal

## Notes

H-T-155 confirmed (same failure as the explicit request, table absent). The bucket-level put does NOT run the
maintenance check itself.

Full original title: Bucket default SSE-KMS with a key lacking the maintenance principal:
PutTableBucketEncryption -> HTTP 200; CreateTable(no encryptionConfiguration) -> BadRequestException 400
