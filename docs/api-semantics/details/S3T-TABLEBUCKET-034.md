<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-034: CreateTableBucket encryption: AES256+kmsKeyArn -> BadRequestException 400...
_Full entry and notes of one finding; its summary entry is in
[tablebucket-encryption-storage.md](../tablebucket-encryption-storage.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-034"></a>**S3T-TABLEBUCKET-034** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **CreateTableBucket encryption: AES256+kmsKeyArn -> BadRequestException 400...**
  With one symmetric CMK (default key policy): {AES256, kmsKeyArn=<key ARN>} -> BadRequestException HTTP 400
  "KMS key ARN detected. Please specify 'aws:kms' as the encryption algorithm."; {aws:kms, kmsKeyArn=<bare key
  id>} -> BadRequestException HTTP 400 "The specified KMS key ARN is invalid. Only key ARN format
  'arn:<partition>:kms:<region>:<account-id>:key/<key-id>' is supported."; {aws:kms, kmsKeyArn=<alias ARN>} ->
  BadRequestException HTTP 400 "The specified KMS key ARN is invalid. Only key ARN format
  'arn:<partition>:kms:<region>:<account-id>:key/<key-id>' is supported."; {aws:kms, kmsKeyArn='alias/...'} ->
  BadRequestException HTTP 400 "The specified KMS key ARN is invalid. Only key ARN format
  'arn:<partition>:kms:<region>:<account-id>:key/<key-id>' is supported."; {aws:kms, kmsKeyArn=<key ARN>} ->
  HTTP 200 and GetTableBucketEncryption returned {"sseAlgorithm": "aws:kms", "kmsKeyArn":
  "arn:aws:kms:us-west-2:<ACCOUNT>:key/12aa39a2-b0c3-475c-aa11-bb7fd4033450"}.
  - ACK: terminal_codes, references, docs-only · ops: CreateTableBucket, GetTableBucketEncryption · fields:
    encryptionConfiguration.sseAlgorithm, encryptionConfiguration.kmsKeyArn
  - repro: CreateTableBucket with each encryptionConfiguration form; GetTableBucketEncryption on any 200
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-035](../tablebucket-encryption-storage.md#s3t-tablebucket-035), [S3T-TABLEBUCKET-181](../tablebucket-encryption-storage.md#s3t-tablebucket-181), [S3T-TABLEBUCKET-106](../tablebucket-policy.md#s3t-tablebucket-106), [S3T-TABLE-086](../table-policy.md#s3t-table-086),
    [S3T-TABLEBUCKET-135](../tablebucket-encryption-storage.md#s3t-tablebucket-135), [S3T-TABLE-258](../table-encryption-storage.md#s3t-table-258), [S3T-TABLEBUCKET-136](../tablebucket-encryption-storage.md#s3t-tablebucket-136), [S3T-TABLEBUCKET-137](../tablebucket-encryption-storage.md#s3t-tablebucket-137), [S3T-TABLEBUCKET-138](../tablebucket-encryption-storage.md#s3t-tablebucket-138),
    [S3T-TABLE-084](../table-encryption-storage.md#s3t-table-084), [S3T-TABLE-085](../table-encryption-storage.md#s3t-table-085) · hypotheses: H-B-012, H-B-013 · evidence:
    tablebucket/sub-resources/kms-encryption-forms

## Notes

H-B-012: confirmed (rejected). H-B-013: confirmed (key id rejected, alias ARN rejected, key ARN echo
identical).

Full original title: CreateTableBucket encryption: AES256+kmsKeyArn -> BadRequestException 400; aws:kms with
key id -> BadRequestException 400, alias ARN -> BadRequestException 400, key ARN -> HTTP 200 (echoed
byte-identical)
