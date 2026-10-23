<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-026: Bucket-core 4xx bodies are restJson {"message"} with the code only in x-amzn-ErrorType; 400 messages name the field class, not the rule
_Full entry and notes of one finding; its summary entry is in [tablebucket.md](../tablebucket.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-tablebucket-026"></a>**S3T-TABLEBUCKET-026** `error-code` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Bucket-core 4xx bodies are restJson {"message"} with the code only in x-amzn-ErrorType; 400 messages name the field class, not the rule**
  Across 63 rejected calls in this probe every error body was {"message": "..."} (no __type member) with
  header x-amzn-ErrorType '<Code>:http://internal.amazon.com/coral/com.amazon.s3tablebuckets/'.
  BadRequestException (HTTP 400) messages identify the field class: 'The specified bucket name isn't
  valid...', 'The specified SSE algorithm is invalid.', 'The specified storage class is not valid.', "The
  specified KMS key ARN is invalid. Only key ARN format
  'arn:<partition>:kms:<region>:<account-id>:key/<key-id>' is supported.", 'The provided tags are invalid.
  Number of user tags must not exceed 50' / 'Tag key is empty or exceeds max length' / 'Tag value exceeds max
  length', "Tags beginning with 'aws:' are reserved for system use.", 'The specified bucket type filter value
  isn't valid...'. NotFoundException (HTTP 404): 'The specified bucket does not exist.' for bucket ARNs; the
  message does not say which ARN segment was wrong.
  - ACK: terminal_codes, docs-only · ops: CreateTableBucket, GetTableBucket, ListTableBuckets
  - repro: Any rejected CreateTableBucket / GetTableBucket; inspect raw HTTP body and x-amzn-ErrorType header
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-023](../tablebucket-encryption-storage.md#s3t-tablebucket-023), [S3T-TABLEBUCKET-024](../tablebucket-encryption-storage.md#s3t-tablebucket-024), [S3T-TABLEBUCKET-104](../tablebucket-encryption-storage.md#s3t-tablebucket-104), [S3T-TABLEBUCKET-087](../tablebucket-encryption-storage.md#s3t-tablebucket-087) · hypotheses:
    H-B-068 · evidence: tablebucket/weird-inputs/create-names-enums-empty-structs

## Notes

Qualifies H-B-068: shape and modeled HTTP statuses confirmed; messages are field-class specific (name / SSE
algorithm / storage class / KMS key ARN / tags) but do not name the violated rule for bucket names. 404
messages are generic as predicted. 403/409/429 bodies are recorded in the idempotency and iam probes.
