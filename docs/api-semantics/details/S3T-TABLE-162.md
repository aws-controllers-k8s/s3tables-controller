<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-162: PutTablePolicy invalid documents: 9/17 variants rejected, all BadRequestException+ParamValidationError...
_Full entry and notes of one finding; its summary entry is in [table-policy.md](../table-policy.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-table-162"></a>**S3T-TABLE-162** `error-code` · impact high · handled · verified 2026-10-13
  **PutTablePolicy invalid documents: 9/17 variants rejected, all BadRequestException+ParamValidationError...**
  Outcome per document (code/HTTP or HTTP on success): {"truncated": "BadRequestException/400",
  "s3:GetObject": "BadRequestException/400", "no Version": "HTTP 200", "NoSuchAction": "HTTP 200",
  "s3tables:*": "HTTP 200", "\"\"": "ParamValidationError/None", "{}": "BadRequestException/400", "duplicate
  Sid": "HTTP 200", "unicode Sid": "HTTP 200", "Version 2008-10-17": "HTTP 200", "Version 2030-01-01":
  "BadRequestException/400", "Effect allow": "HTTP 200", "no Statement": "BadRequestException/400", "Statement
  []": "BadRequestException/400", "unknown keys Id/Foo": "BadRequestException/400", "no Principal":
  "BadRequestException/400", "NotPrincipal Deny": "HTTP 200"}. Messages: {"truncated": "The specified policy
  is malformed", "s3:GetObject": "Invalid Action", "\"\"": "Parameter validation failed:\nInvalid length for
  parameter resourcePolicy, value: 0, valid min length: 1", "{}": "The specified policy is malformed",
  "Version 2030-01-01": "The specified policy is malformed", "no Statement": "The specified policy is
  malformed", "Statement []": "The specified policy is malformed", "no Principal": "Policy contains a
  statement with no principals"}. Variants that were accepted are read back exactly as recorded in result.yaml
  (case_* entries).
  - ACK: terminal_codes, is_iam_policy · ops: PutTablePolicy · fields: resourcePolicy
  - repro: PutTablePolicy on a fresh table with each document variant; Get+Delete after each accepted one
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLE-157](../table-policy.md#s3t-table-157), [S3T-TABLE-164](../table-policy.md#s3t-table-164), [S3T-TABLE-166](../table-policy.md#s3t-table-166), [S3T-TABLE-167](../table-policy.md#s3t-table-167), [S3T-TABLEPOLICY-001](../table-policy.md#s3t-tablepolicy-001),
    [S3T-TABLEBUCKET-088](../tablebucket-policy.md#s3t-tablebucket-088), [S3T-TABLEPOLICY-009](../table-policy.md#s3t-tablepolicy-009), [S3T-TABLEBUCKETPOLICY-012](../tablebucket-policy.md#s3t-tablebucketpolicy-012) · hypotheses: H-U-008, H-U-027 ·
    evidence: table/error-taxonomy/policy-validation

## Notes

H-U-008 qualified (truncated=BadRequestException/400, s3:GetObject=BadRequestException/400, no Version=HTTP
200, NoSuchAction=HTTP 200; distinct messages=2). H-U-027: empty string=ParamValidationError/None,
{}=BadRequestException/400, duplicate Sid=HTTP 200, unicode Sid=HTTP 200, Version 2008=HTTP 200.

Full original title: PutTablePolicy invalid documents: 9/17 variants rejected, all
BadRequestException+ParamValidationError; truncated JSON, s3:GetObject, missing Version ->
BadRequestException/400 / BadRequestException/400 / HTTP 200
