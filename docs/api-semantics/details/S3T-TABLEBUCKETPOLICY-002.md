<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKETPOLICY-002: PutTableBucketPolicy size limit: 20480 bytes accepted, 20481 bytes rejected with ValidationException (HTTP 400)
_Full entry and notes of one finding; its summary entry is in
[tablebucket-policy.md](../tablebucket-policy.md). Generated from aws-api-semantics `services/s3tables` (model
2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucketpolicy-002"></a>**S3T-TABLEBUCKETPOLICY-002** `quota-limit` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **PutTableBucketPolicy size limit: 20480 bytes accepted, 20481 bytes rejected with ValidationException (HTTP 400)**
  Valid Allow policies padded with unique Sids: 19000 bytes (96 statements) -> HTTP 200; 20000 bytes (101
  statements) -> HTTP 200; 20480 bytes (103 statements) -> HTTP 200; 20481 bytes (103 statements) ->
  ValidationException (HTTP 400); 21000 bytes (106 statements) -> ValidationException (HTTP 400). Bisection to
  the byte: largest accepted 20480, smallest rejected 20481. Rejection message: '1 validation error detected:
  Value at 'resourcePolicy' failed to satisfy constraint: Member must have length less than or equal to
  20480'. GetTableBucketPolicy after the accepted puts returns 20480 bytes (the stored document is
  byte-identical in length to the sent one).
  - ACK: terminal_codes, is_iam_policy · ops: PutTableBucketPolicy, GetTableBucketPolicy · fields:
    resourcePolicy
  - repro: PutTableBucketPolicy with ASCII policies of 19000/20000/20480/20481/21000 bytes (client validation
    off), then bisect
  - measurements: largest_accepted_bytes=20480, smallest_rejected_bytes=20481, n=5
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-088](../tablebucket-policy.md#s3t-tablebucket-088), [S3T-TABLEBUCKET-089](../tablebucket-policy.md#s3t-tablebucket-089), [S3T-TABLEBUCKET-090](../tablebucket-policy.md#s3t-tablebucket-090), [S3T-TABLEBUCKETPOLICY-001](../tablebucket-policy.md#s3t-tablebucketpolicy-001),
    [S3T-TABLEBUCKETPOLICY-014](../tablebucket-policy.md#s3t-tablebucketpolicy-014), [S3T-TABLEBUCKETPOLICY-012](../tablebucket-policy.md#s3t-tablebucketpolicy-012), [S3T-TABLEPOLICY-001](../table-policy.md#s3t-tablepolicy-001), [S3T-TABLE-120](../table.md#s3t-table-120) · hypotheses:
    H-S-019 · evidence: tablebucket/limits/read-write-bursts-policy-size

## Notes

H-S-019 qualified: a size limit exists but at exactly 20480 bytes (the SDK model max), not 20 KB = 20000;
19000/20000/20480 accepted, 20481 and 21000 rejected with wire code ValidationException (HTTP 400) - not
BadRequestException and not in the modeled error list - '1 validation error detected: Value at
'resourcePolicy' failed to satisfy constraint: Member must have length less than or equal to 20480'.
GetTableBucketPolicy returned the 20480-byte document with the same length. Counted in ASCII bytes ==
characters here.
