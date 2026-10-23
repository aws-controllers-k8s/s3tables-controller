<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-103: PutTableBucketEncryption kmsKeyArn: bare key id, alias name, alias ARN, upper-case region/prefix -> 400; key ARN echoed byte-identical
_Full entry and notes of one finding; its summary entry is in
[tablebucket-encryption-storage.md](../tablebucket-encryption-storage.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-103"></a>**S3T-TABLEBUCKET-103** `normalization` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **PutTableBucketEncryption kmsKeyArn: bare key id, alias name, alias ARN, upper-case region/prefix -> 400; key ARN echoed byte-identical**
  key_id -> BadRequestException 400 ('The specified KMS key ARN is invalid. Only key ARN format
  'arn:<partition>:kms:<region>:<account-id>:key/<key-id>' is supported.'); alias_name -> BadRequestException
  400 ('The specified KMS key ARN is invalid. Only key ARN format
  'arn:<partition>:kms:<region>:<account-id>:key/<key-id>' is supported.'); alias_arn -> BadRequestException
  400 ('The specified KMS key ARN is invalid. Only key ARN format
  'arn:<partition>:kms:<region>:<account-id>:key/<key-id>' is supported.'); upper_region ->
  BadRequestException 400 ('The specified KMS key ARN is invalid. Only key ARN format
  'arn:<partition>:kms:<region>:<account-id>:key/<key-id>' is supported.'); upper_prefix ->
  BadRequestException 400 ('The specified KMS key ARN is invalid. Only key ARN format
  'arn:<partition>:kms:<region>:<account-id>:key/<key-id>' is supported.'). The full key ARN -> HTTP 200 and
  Get returned the kmsKeyArn byte-identical to the sent string.
  - ACK: terminal_codes, references · ops: PutTableBucketEncryption, GetTableBucketEncryption · fields:
    encryptionConfiguration.kmsKeyArn
  - repro: Put aws:kms with kmsKeyArn = key id / alias name / alias ARN / ARN with upper-case region / ARN
    with upper-case prefix / key ARN; Get after each
  - measurements: n=6
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-050](../tablebucket-encryption-storage.md#s3t-tablebucket-050), [S3T-TABLEBUCKET-052](../tablebucket-encryption-storage.md#s3t-tablebucket-052), [S3T-TABLEBUCKET-083](../tablebucket-encryption-storage.md#s3t-tablebucket-083), [S3T-TABLE-059](../table-encryption-storage.md#s3t-table-059),
    [S3T-TABLEBUCKET-102](../tablebucket-encryption-storage.md#s3t-tablebucket-102), [S3T-TABLEBUCKET-023](../tablebucket-encryption-storage.md#s3t-tablebucket-023), [S3T-TABLE-084](../table-encryption-storage.md#s3t-table-084), [S3T-TABLE-085](../table-encryption-storage.md#s3t-table-085) · hypotheses: H-S-005 · evidence:
    tablebucket/weird-inputs/encryption-put-validation

## Notes

H-S-005 confirmed: bare key id, alias name, alias ARN, ARN with upper-cased region and ARN with upper-cased
'ARN:AWS:KMS' prefix are all rejected with BadRequestException 400 "The specified KMS key ARN is invalid. Only
key ARN format 'arn:<partition>:kms:<region>:<account-id>:key/<key-id>' is supported."; the full key ARN is
echoed byte-identical by Get (no alias resolution).

Full original title: PutTableBucketEncryption kmsKeyArn forms: bare key id -> BadRequestException 400, alias
name -> BadRequestException 400, alias ARN -> BadRequestException 400, upper-case region ->
BadRequestException 400; key ARN echoed byte-identical
