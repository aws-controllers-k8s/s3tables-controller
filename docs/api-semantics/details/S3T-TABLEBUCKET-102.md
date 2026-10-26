<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-102: PutTableBucketEncryption: aws:kms without kmsKeyArn or with '' -> 400; AES256 with any kmsKeyArn (even '') -> 400; stored value unchanged
_Full entry and notes of one finding; its summary entry is in
[tablebucket-encryption-storage.md](../tablebucket-encryption-storage.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-102"></a>**S3T-TABLEBUCKET-102** `request-validation` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **PutTableBucketEncryption: aws:kms without kmsKeyArn or with '' -> 400; AES256 with any kmsKeyArn (even '') -> 400; stored value unchanged**
  Baseline aws:kms+key stored. kms_no_key -> BadRequestException 400 ('The specified KMS key ARN is invalid.
  Only key ARN format 'arn:<partition>:kms:<region>:<account-id>:key/<key-id>' is supported.'), Get after ->
  aws:kms+key; aes_with_key -> BadRequestException 400 ('KMS key ARN detected. Please specify 'aws:kms' as the
  encryption algorithm.'), Get after -> aws:kms+key; kms_empty_key -> BadRequestException 400 ('The specified
  KMS key ARN is invalid. Only key ARN format 'arn:<partition>:kms:<region>:<account-id>:key/<key-id>' is
  supported.'), Get after -> aws:kms+key; aes_empty_key -> BadRequestException 400 ('KMS key ARN detected.
  Please specify 'aws:kms' as the encryption algorithm.'), Get after -> aws:kms+key. Get after each rejected
  call still returned the baseline (True).
  - ACK: terminal_codes, docs-only · ops: PutTableBucketEncryption, GetTableBucketEncryption · fields:
    encryptionConfiguration.sseAlgorithm, encryptionConfiguration.kmsKeyArn
  - repro: Put aws:kms+key; then Put {aws:kms}, {AES256,key}, {aws:kms,''}, {AES256,''} with SDK validation
    off; Get after each
  - measurements: n=4
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-103](../tablebucket-encryption-storage.md#s3t-tablebucket-103), [S3T-TABLEBUCKET-023](../tablebucket-encryption-storage.md#s3t-tablebucket-023), [S3T-TABLE-084](../table-encryption-storage.md#s3t-table-084), [S3T-TABLE-085](../table-encryption-storage.md#s3t-table-085) · hypotheses: H-S-004 ·
    evidence: tablebucket/weird-inputs/encryption-put-validation

## Notes

H-S-004 confirmed: aws:kms without kmsKeyArn and aws:kms with '' -> 400 'The specified KMS key ARN is invalid.
Only key ARN format ... is supported.'; AES256 with a key ARN and AES256 with '' -> 400 "KMS key ARN detected.
Please specify 'aws:kms' as the encryption algorithm." Note: an empty string kmsKeyArn is treated as
'detected' for AES256 and as 'invalid' for aws:kms; the controller must omit the member, not send ''.

Full original title: PutTableBucketEncryption: aws:kms without kmsKeyArn -> BadRequestException 400; AES256
with kmsKeyArn -> BadRequestException 400; aws:kms with kmsKeyArn '' -> BadRequestException 400; AES256 with
'' -> BadRequestException 400
