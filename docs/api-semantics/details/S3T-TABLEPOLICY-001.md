<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEPOLICY-001: PutTablePolicy size limit: 20480 bytes accepted, 20481 bytes rejected with ValidationException (HTTP 400)
_Full entry and notes of one finding; its summary entry is in [table-policy.md](../table-policy.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-tablepolicy-001"></a>**S3T-TABLEPOLICY-001** `quota-limit` · impact medium · handled · verified 2026-10-13
  **PutTablePolicy size limit: 20480 bytes accepted, 20481 bytes rejected with ValidationException (HTTP 400)**
  Valid Allow policies padded with unique Sids on one table: 16000 bytes (68 statements) -> HTTP 200; 19500
  bytes (83 statements) -> HTTP 200; 20480 bytes (87 statements) -> HTTP 200; 20481 bytes (87 statements) ->
  ValidationException (HTTP 400); 32000 bytes (136 statements) -> ValidationException (HTTP 400); 65536 bytes
  (279 statements) -> ValidationException (HTTP 400). Rejection message: '1 validation error detected: Value
  at 'resourcePolicy' failed to satisfy constraint: Member must have length less than or equal to 20480'.
  GetTablePolicy afterwards returned 20480 bytes (the last accepted document was 20480 bytes).
  - ACK: terminal_codes, is_iam_policy · ops: PutTablePolicy, GetTablePolicy · fields: resourcePolicy
  - repro: PutTablePolicy with ASCII policies of 16000/19500/20480/20481/32000/65536 bytes (client validation
    off)
  - measurements: largest_accepted_bytes=20480, smallest_rejected_bytes=20481, n=6
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLEBUCKETPOLICY-002](../tablebucket-policy.md#s3t-tablebucketpolicy-002), [S3T-TABLEBUCKETPOLICY-014](../tablebucket-policy.md#s3t-tablebucketpolicy-014), [S3T-TABLE-120](../table.md#s3t-table-120), [S3T-TABLEBUCKET-090](../tablebucket-policy.md#s3t-tablebucket-090),
    [S3T-TABLE-157](../table-policy.md#s3t-table-157), [S3T-TABLE-162](../table-policy.md#s3t-table-162), [S3T-TABLE-164](../table-policy.md#s3t-table-164), [S3T-TABLE-166](../table-policy.md#s3t-table-166), [S3T-TABLE-167](../table-policy.md#s3t-table-167) · hypotheses: H-U-011 ·
    evidence: table/limits/policy-size-write-races

## Notes

H-U-011 qualified: a hard limit exists but at exactly 20480 bytes (the SDK model max for resourcePolicy), not
~20 KB = 20000: 16000/19500/20480 accepted, 20481/32000/65536 rejected with wire code ValidationException
(HTTP 400) - not BadRequestException, not in the modeled error list - '1 validation error detected: Value at
'resourcePolicy' failed to satisfy constraint: Member must have length less than or equal to 20480'. Same
boundary and code as PutTableBucketPolicy. botocore does not enforce the max client-side.
