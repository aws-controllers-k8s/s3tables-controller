<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-088: PutTableBucketPolicy: bad JSON, non-s3tables action, missing Principal -> 400; unknown s3tables:* action, other bucket's Resource -> 200
_Full entry and notes of one finding; its summary entry is in
[tablebucket-policy.md](../tablebucket-policy.md). Generated from aws-api-semantics `services/s3tables` (model
2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-088"></a>**S3T-TABLEBUCKET-088** `request-validation` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **PutTableBucketPolicy: bad JSON, non-s3tables action, missing Principal -> 400; unknown s3tables:* action, other bucket's Resource -> 200**
  With a valid baseline policy stored: not_json -> BadRequestException 400; unknown_namespace ->
  BadRequestException 400; unknown_s3tables_action -> HTTP 200 (STORED); s3_action -> BadRequestException 400;
  foreign_resource -> HTTP 200 (STORED); no_principal -> BadRequestException 400. Messages: not_json: 'The
  specified policy is malformed'; unknown_namespace: 'Invalid Action'; s3_action: 'Invalid Action';
  no_principal: 'Policy contains a statement with no principals'. 'STORED' marks documents that replaced the
  baseline (Get showed them).
  - ACK: terminal_codes, is_iam_policy · ops: PutTableBucketPolicy, GetTableBucketPolicy · fields:
    resourcePolicy
  - repro: Put baseline; for each document: Put -> record -> Get -> compare with baseline
  - measurements: n=6
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-089](../tablebucket-policy.md#s3t-tablebucket-089), [S3T-TABLEBUCKET-090](../tablebucket-policy.md#s3t-tablebucket-090), [S3T-TABLEBUCKETPOLICY-001](../tablebucket-policy.md#s3t-tablebucketpolicy-001), [S3T-TABLEBUCKETPOLICY-002](../tablebucket-policy.md#s3t-tablebucketpolicy-002),
    [S3T-TABLEBUCKETPOLICY-014](../tablebucket-policy.md#s3t-tablebucketpolicy-014), [S3T-TABLEBUCKETPOLICY-012](../tablebucket-policy.md#s3t-tablebucketpolicy-012), [S3T-TABLE-167](../table-policy.md#s3t-table-167), [S3T-TABLE-162](../table-policy.md#s3t-table-162), [S3T-TABLEPOLICY-009](../table-policy.md#s3t-tablepolicy-009) ·
    hypotheses: H-S-016 · evidence: tablebucket/weird-inputs/policy-validation-lockout

## Notes

H-S-016 qualified: (a) bad JSON -> BadRequestException 400 'The specified policy is malformed'; (b) 'foo:Bar'
and 's3:GetObject' -> BadRequestException 400 'Invalid Action' (namespace must be s3tables), yet
's3tables:NotAnAction' -> HTTP 200 and is stored; (c) REFUTED: Resource = another existing table bucket's
ARN/table/* -> HTTP 200, stored as sent (no resource-must-match-bucket rule); (d) no Principal ->
BadRequestException 400 'Policy contains a statement with no principals'.

Full original title: PutTableBucketPolicy rejects bad JSON, foreign action namespaces and missing Principal
(400) but ACCEPTS unknown s3tables:* actions and another bucket's Resource ARN
