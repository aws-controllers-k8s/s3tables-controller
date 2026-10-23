<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-087: CreateTable sseAlgorithm is case-sensitive: 'aes256', 'Aes256', 'AES-256', 'Aws:Kms', 'AWS:KMS', 'aws-kms', 'kms', 'aws:kms:dsse' -> 400
_Full entry and notes of one finding; its summary entry is in
[table-encryption-storage.md](../table-encryption-storage.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-087"></a>**S3T-TABLE-087** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **CreateTable sseAlgorithm is case-sensitive: 'aes256', 'Aes256', 'AES-256', 'Aws:Kms', 'AWS:KMS', 'aws-kms', 'kms', 'aws:kms:dsse' -> 400**
  'aes256' -> BadRequestException (HTTP 400) 'The specified SSE algorithm is invalid.'; 'Aes256' ->
  BadRequestException (HTTP 400) 'The specified SSE algorithm is invalid.'; 'AES-256' -> BadRequestException
  (HTTP 400) 'The specified SSE algorithm is invalid.'; 'Aws:Kms' -> BadRequestException (HTTP 400) 'The
  specified SSE algorithm is invalid.'; 'AWS:KMS' -> BadRequestException (HTTP 400) 'The specified SSE
  algorithm is invalid.'; 'aws-kms' -> BadRequestException (HTTP 400) 'The specified SSE algorithm is
  invalid.'; 'aws:kms ' -> BadRequestException (HTTP 400) 'The specified SSE algorithm is invalid.'; 'kms' ->
  BadRequestException (HTTP 400) 'The specified SSE algorithm is invalid.'; 'aws:kms:dsse' ->
  BadRequestException (HTTP 400) 'The specified SSE algorithm is invalid.'. Distinct messages: ['The specified
  SSE algorithm is invalid.']. (Exact 'AES256' and 'aws:kms' succeed in the sibling probes.)
  - ACK: terminal_codes · ops: CreateTable · fields: encryptionConfiguration.sseAlgorithm
  - repro: CreateTable(encryptionConfiguration={sseAlgorithm: <variant>[, kmsKeyArn: K]}) per variant,
    distinct names
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-024](../tablebucket-encryption-storage.md#s3t-tablebucket-024), [S3T-TABLEBUCKET-104](../tablebucket-encryption-storage.md#s3t-tablebucket-104), [S3T-TABLEBUCKET-023](../tablebucket-encryption-storage.md#s3t-tablebucket-023), [S3T-TABLE-013](../table-encryption-storage.md#s3t-table-013), [S3T-TABLE-012](../table-encryption-storage.md#s3t-table-012),
    [S3T-TABLE-084](../table-encryption-storage.md#s3t-table-084), [S3T-TABLE-085](../table-encryption-storage.md#s3t-table-085) · hypotheses: H-T-181 · evidence: table/weird-inputs/encryption-storage-enums

## Notes

H-T-181 confirmed. The message does not name the enum values.

Full original title: CreateTable sseAlgorithm is case-sensitive: 'aes256', 'Aes256', 'AES-256', 'Aws:Kms',
'AWS:KMS', 'aws-kms', 'aws:kms ', 'kms', 'aws:kms:dsse' -> all rejected BadRequestException/400; only exact
'AES256' / 'aws:kms' succeed
