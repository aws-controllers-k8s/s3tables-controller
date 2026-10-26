<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-085: CreateTable kmsKeyArn: only the exact full key ARN accepted; bare id, alias name/ARN, uppercased or trailing-space ARN -> 400
_Full entry and notes of one finding; its summary entry is in
[table-encryption-storage.md](../table-encryption-storage.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-085"></a>**S3T-TABLE-085** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **CreateTable kmsKeyArn: only the exact full key ARN accepted; bare id, alias name/ARN, uppercased or trailing-space ARN -> 400**
  bare key id -> BadRequestException (HTTP 400) 'The specified KMS key ARN is invalid. Only key ARN format
  'arn:<partition>:kms:<region>:<account-id>:key/<key-id>' is supported.'; alias name -> BadRequestException
  (HTTP 400) 'The specified KMS key ARN is invalid. Only key ARN format
  'arn:<partition>:kms:<region>:<account-id>:key/<key-id>' is supported.'; alias ARN -> BadRequestException
  (HTTP 400) 'The specified KMS key ARN is invalid. Only key ARN format
  'arn:<partition>:kms:<region>:<account-id>:key/<key-id>' is supported.'; full key ARN -> HTTP 200,
  GetTableEncryption {"sseAlgorithm": "aws:kms", "kmsKeyArn": "<K ARN>"}; full key ARN + trailing space ->
  BadRequestException (HTTP 400) 'The specified KMS key ARN does not exist.'; full key ARN uppercased ->
  BadRequestException (HTTP 400) 'The specified KMS key ARN is invalid. Only key ARN format
  'arn:<partition>:kms:<region>:<account-id>:key/<key-id>' is supported.'; key ARN with 'kms' service
  uppercased -> BadRequestException (HTTP 400) 'The specified KMS key ARN is invalid. Only key ARN format
  'arn:<partition>:kms:<region>:<account-id>:key/<key-id>' is supported.'.
  - ACK: terminal_codes, references · ops: CreateTable, GetTableEncryption · fields:
    encryptionConfiguration.kmsKeyArn
  - repro: CreateTable(aws:kms, kmsKeyArn=<form>) for each form, distinct names; GetTableEncryption on success
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-102](../tablebucket-encryption-storage.md#s3t-tablebucket-102), [S3T-TABLEBUCKET-103](../tablebucket-encryption-storage.md#s3t-tablebucket-103), [S3T-TABLEBUCKET-023](../tablebucket-encryption-storage.md#s3t-tablebucket-023), [S3T-TABLE-084](../table-encryption-storage.md#s3t-table-084),
    [S3T-TABLEBUCKET-105](../tablebucket-encryption-storage.md#s3t-tablebucket-105), [S3T-TABLEBUCKET-128](../tablebucket-encryption-storage.md#s3t-tablebucket-128), [S3T-TABLEBUCKET-107](../tablebucket-encryption-storage.md#s3t-tablebucket-107), [S3T-TABLEBUCKET-131](../tablebucket.md#s3t-tablebucket-131), [S3T-TABLEBUCKET-135](../tablebucket-encryption-storage.md#s3t-tablebucket-135),
    [S3T-TABLE-258](../table-encryption-storage.md#s3t-table-258), [S3T-TABLEBUCKET-136](../tablebucket-encryption-storage.md#s3t-tablebucket-136), [S3T-TABLEBUCKET-137](../tablebucket-encryption-storage.md#s3t-tablebucket-137), [S3T-TABLEBUCKET-138](../tablebucket-encryption-storage.md#s3t-tablebucket-138), [S3T-TABLEBUCKET-034](../tablebucket-encryption-storage.md#s3t-tablebucket-034),
    [S3T-TABLE-012](../table-encryption-storage.md#s3t-table-012), [S3T-TABLE-013](../table-encryption-storage.md#s3t-table-013), [S3T-TABLE-087](../table-encryption-storage.md#s3t-table-087) · hypotheses: H-T-158, H-T-159 · evidence:
    table/weird-inputs/encryption-storage-enums

## Notes

H-T-158 confirmed (only the full key ARN is accepted). H-T-159 (alias ARN accepted and resolved to the key
ARN) refuted (alias ARN rejected).

Full original title: CreateTable kmsKeyArn forms: accepted full key ARN; rejected bare key id
(BadRequestException 400), alias name (BadRequestException 400), alias ARN (BadRequestException 400), full key
ARN + trailing space (BadRequestException 400), full key ARN uppercased (BadRequestException 400), key ARN
with 'kms' service uppercased (BadRequestException 400)
