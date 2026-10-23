<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-090: PutTableBucketPolicy with resourcePolicy '' -> ValidationException 400, whitespace -> BadRequestException 400...
_Full entry and notes of one finding; its summary entry is in
[tablebucket-policy.md](../tablebucket-policy.md). Generated from aws-api-semantics `services/s3tables` (model
2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-090"></a>**S3T-TABLEBUCKET-090** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **PutTableBucketPolicy with resourcePolicy '' -> ValidationException 400, whitespace -> BadRequestException 400...**
  SDK validation off. '' -> ValidationException 400 ('1 validation error detected: Value at 'resourcePolicy'
  failed to satisfy constraint: Member must have length greater than or equal to 1'); ' ' ->
  BadRequestException 400 ('The specified policy is malformed'); '{}' -> BadRequestException 400 ('The
  specified policy is malformed'); BOM-prefixed valid document -> BadRequestException 400; Sid 'ÄckÜnicode-日本'
  -> HTTP 200 (Sid returned unchanged). After each rejected Put the baseline policy was still stored (True);
  none acted as a delete.
  - ACK: custom_delete, terminal_codes · ops: PutTableBucketPolicy, GetTableBucketPolicy · fields:
    resourcePolicy
  - repro: Put baseline; Put '' / ' ' / '{}' / BOM+valid / unicode Sid with client validation off; Get after
    each
  - measurements: n=5
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-088](../tablebucket-policy.md#s3t-tablebucket-088), [S3T-TABLEBUCKET-089](../tablebucket-policy.md#s3t-tablebucket-089), [S3T-TABLEBUCKETPOLICY-001](../tablebucket-policy.md#s3t-tablebucketpolicy-001), [S3T-TABLEBUCKETPOLICY-002](../tablebucket-policy.md#s3t-tablebucketpolicy-002),
    [S3T-TABLEBUCKETPOLICY-014](../tablebucket-policy.md#s3t-tablebucketpolicy-014), [S3T-TABLEBUCKETPOLICY-012](../tablebucket-policy.md#s3t-tablebucketpolicy-012), [S3T-TABLE-120](../table.md#s3t-table-120), [S3T-TABLEPOLICY-001](../table-policy.md#s3t-tablepolicy-001) · hypotheses:
    H-S-024 · evidence: tablebucket/weird-inputs/policy-validation-lockout

## Notes

H-S-024 qualified: '' -> ValidationException 400 (unmodeled wire code, min length 1); whitespace-only and '{}'
-> BadRequestException 400 'The specified policy is malformed' (none acted as a delete: confirmed); a UTF-8
BOM before a valid document is REJECTED (BadRequestException 400, hypothesis expected accepted); the non-ASCII
Sid was accepted and returned unchanged (confirmed).

Full original title: PutTableBucketPolicy with resourcePolicy '' -> ValidationException 400, whitespace ->
BadRequestException 400, '{}' -> BadRequestException 400, UTF-8 BOM -> BadRequestException 400, unicode Sid ->
HTTP 200
