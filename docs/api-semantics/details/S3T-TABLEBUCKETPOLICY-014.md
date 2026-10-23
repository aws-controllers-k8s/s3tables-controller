<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKETPOLICY-014: PutTableBucketPolicy 20480-byte limit is applied to the RAW bytes as sent (whitespace counts)...
_Full entry and notes of one finding; its summary entry is in
[tablebucket-policy.md](../tablebucket-policy.md). Generated from aws-api-semantics `services/s3tables` (model
2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucketpolicy-014"></a>**S3T-TABLEBUCKETPOLICY-014** `quota-limit` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **PutTableBucketPolicy 20480-byte limit is applied to the RAW bytes as sent (whitespace counts)...**
  Same 60-statement document: minified (12098 B) -> HTTP 200 (read-back 12098 B); pretty-printed and
  space-padded to 20479 B -> HTTP 200; to 20480 B -> HTTP 200; to 20481 B -> ValidationException (HTTP 400) '1
  validation error detected: Value at 'resourcePolicy' failed to satisfy constraint: Member must have length
  less than or equal to 20480'; to 30000 B -> ValidationException (HTTP 400) '1 validation error detected:
  Value at 'resourcePolicy' failed to satisfy constraint: Member must have length less than or equal to
  20480'. Control: a minified 102-statement document of 20540 B -> ValidationException (HTTP 400) '1
  validation error detected: Value at 'resourcePolicy' failed to satisfy constraint: Member must have length
  less than or equal to 20480'.
  - ACK: is_iam_policy, terminal_codes · ops: PutTableBucketPolicy, GetTableBucketPolicy · fields:
    resourcePolicy
  - repro: Put the same policy minified and pretty-printed/padded to 20479, 20480, 20481 and 30000 bytes
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETPOLICY-002](../tablebucket-policy.md#s3t-tablebucketpolicy-002), [S3T-TABLEBUCKET-063](../tablebucket-policy.md#s3t-tablebucket-063), [S3T-TABLEBUCKET-088](../tablebucket-policy.md#s3t-tablebucket-088), [S3T-TABLEBUCKET-089](../tablebucket-policy.md#s3t-tablebucket-089),
    [S3T-TABLEBUCKET-090](../tablebucket-policy.md#s3t-tablebucket-090), [S3T-TABLEBUCKETPOLICY-001](../tablebucket-policy.md#s3t-tablebucketpolicy-001), [S3T-TABLEBUCKETPOLICY-012](../tablebucket-policy.md#s3t-tablebucketpolicy-012), [S3T-TABLEPOLICY-001](../table-policy.md#s3t-tablepolicy-001) · evidence:
    tablebucketpolicy/creative/size-limit-whitespace

## Notes

Extends [S3T-TABLEBUCKETPOLICY-002](../tablebucket-policy.md#s3t-tablebucketpolicy-002) (20480/20481 boundary) with WHAT is measured. Whitespace counts: a spec
document that the user pretty-prints can exceed the limit while the service's own minified read-back of the
same policy is far below it, so a controller that round-trips the read-back would succeed where the user's
spec fails - and a 'why does my 15 KB policy fail' report has a 400 ValidationException naming the limit;
minifying before Put is a safe normalization. Single run.

Full original title: PutTableBucketPolicy 20480-byte limit is applied to the RAW bytes as sent (whitespace
counts): a 60-statement policy (12098 B minified) is rejected when pretty-printed to 20481 B
