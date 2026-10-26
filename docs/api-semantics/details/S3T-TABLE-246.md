<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-246: CreateTable(aws:kms) with an RSA_2048 or HMAC_256 key -> BadRequestException 400 maintenance-principal text; bucket Put accepts them
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-246"></a>**S3T-TABLE-246** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **CreateTable(aws:kms) with an RSA_2048 or HMAC_256 key -> BadRequestException 400 maintenance-principal text; bucket Put accepts them**
  Three Enabled keys with the same policy (account root + maintenance.s3tables.amazonaws.com
  kms:GenerateDataKey/kms:Decrypt). CreateTable(format=ICEBERG, encryptionConfiguration={aws:kms, kmsKeyArn})
  with K2 (RSA_2048, SIGN_VERIFY) and with K3 (HMAC_256, GENERATE_VERIFY_MAC) -> BadRequestException HTTP 400
  'Insufficient access to perform table maintenance. Ensure that the key used to encrypt the table still
  exists, is active, and has a resource policy granting access to the S3 service principal
  'maintenance.s3tables.amazonaws.com'.' (x-amzn-ErrorType
  BadRequestException:http://internal.amazon.com/coral/com.amazon.s3tablebuckets/; identical on a second
  attempt 10 s later; no table created). The same call with K1 (SYMMETRIC_DEFAULT) -> HTTP 200 and
  GetTableEncryption echoes {sseAlgorithm: aws:kms, kmsKeyArn: <K1 ARN>}. PutTableBucketEncryption(aws:kms K2)
  -> HTTP 200 and (aws:kms K3) -> HTTP 200 (GetTableBucketEncryption echoes the RSA/HMAC key ARN, no key-type
  validation at the bucket level); a table then created WITHOUT encryptionConfiguration under that bucket
  default fails with the same BadRequestException 400 maintenance text.
  - ACK: terminal_codes · ops: CreateTable, PutTableBucketEncryption · fields:
    encryptionConfiguration.kmsKeyArn, encryptionConfiguration.sseAlgorithm
  - repro: CreateKey(RSA_2048 SIGN_VERIFY), CreateKey(HMAC_256) with the maintenance-principal statement ->
    CreateTable(aws:kms, each key); PutTableBucketEncryption(aws:kms, each key)
  - measurements: create_table_latency_ms.rsa=414, create_table_latency_ms.hmac=411,
    create_table_latency_ms.symmetric=652, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-253](../table-maintenance.md#s3t-table-253), [S3T-TABLE-301](../table-maintenance.md#s3t-table-301), [S3T-TABLE-254](../table-maintenance.md#s3t-table-254), [S3T-TABLE-248](../table-encryption-storage.md#s3t-table-248), [S3T-TABLE-054](../table-maintenance.md#s3t-table-054), [S3T-TABLE-239](../table-record-expiration.md#s3t-table-239),
    [S3T-TABLE-055](../table-maintenance.md#s3t-table-055), [S3T-TABLEBUCKET-119](../tablebucket-encryption-storage.md#s3t-tablebucket-119) · hypotheses: H-T-188 · evidence:
    table/error-taxonomy/kms-key-states-types

## Notes

H-T-188 qualified: asymmetric and HMAC keys are rejected with BadRequestException 400 as predicted, but the
message is the generic maintenance-principal text, not a KMS 'key usage' text - a user cannot tell a wrong key
type from a missing key-policy statement. The multi-region (mrk-) half of H-T-188 was NOT tested (three-key
budget). Bucket-level default encryption does not validate the key type, so the error only surfaces at
CreateTable.

Full original title: CreateTable(aws:kms) with an RSA_2048 SIGN_VERIFY or HMAC_256 key -> BadRequestException
400 with the maintenance-principal text; PutTableBucketEncryption accepts both keys
