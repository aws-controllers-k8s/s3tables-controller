<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-164: PutTablePolicy Principal: nonexistent role ARN BadRequestException/400; sts assumed-role session ARN HTTP 200...
_Full entry and notes of one finding; its summary entry is in [table-policy.md](../table-policy.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-table-164"></a>**S3T-TABLE-164** `request-validation` · impact medium · handled · verified 2026-10-13
  **PutTablePolicy Principal: nonexistent role ARN BadRequestException/400; sts assumed-role session ARN HTTP 200...**
  Outcome per Principal value: {"sts assumed-role session ARN": "HTTP 200", "Service example.com":
  "BadRequestException/400", "Service s3.amazonaws.com": "HTTP 200", "nonexistent role ARN (own acct)":
  "BadRequestException/400", "own :root ARN": "HTTP 200", "nonexistent user ARN": "BadRequestException/400",
  "Federated cognito": "BadRequestException/400", "AWS [role, *]": "BadRequestException/400"}. Messages:
  {"Service example.com": "Invalid service principal in policy", "nonexistent role ARN (own acct)": "Invalid
  principal in policy", "nonexistent user ARN": "Invalid principal in policy", "AWS [role, *]": "Public
  policies cannot be applied to S3 table buckets and tables."}. Read-backs of accepted forms are in
  result.yaml (case_principal_*).
  - ACK: terminal_codes, is_iam_policy · ops: PutTablePolicy · fields: resourcePolicy
  - repro: PutTablePolicy on the fixture table with each Principal form
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-SERVICE-010](../service.md#s3t-service-010), [S3T-SERVICE-025](../service.md#s3t-service-025), [S3T-SERVICE-012](../service.md#s3t-service-012), [S3T-TABLEBUCKETPOLICY-001](../tablebucket-policy.md#s3t-tablebucketpolicy-001),
    [S3T-TABLEBUCKETPOLICY-013](../tablebucket-policy.md#s3t-tablebucketpolicy-013), [S3T-TABLEPOLICY-002](../table-policy.md#s3t-tablepolicy-002), [S3T-TABLE-165](../table-policy.md#s3t-table-165), [S3T-TABLEPOLICY-004](../table-policy.md#s3t-tablepolicy-004), [S3T-TABLE-157](../table-policy.md#s3t-table-157),
    [S3T-TABLE-162](../table-policy.md#s3t-table-162), [S3T-TABLE-166](../table-policy.md#s3t-table-166), [S3T-TABLE-167](../table-policy.md#s3t-table-167), [S3T-TABLEPOLICY-001](../table-policy.md#s3t-tablepolicy-001) · hypotheses: H-U-010, H-U-068 ·
    evidence: table/error-taxonomy/policy-validation

## Notes

H-U-010 first half (nonexistent role rejected): confirmed. H-U-068 principal part: sts session ARN ACCEPTED
(readback
{"Version":"2012-10-17","Statement":[{"Sid":"X","Effect":"Allow","Principal":{"AWS":"arn:aws:sts::<ACCOUNT>:assumed-r),
non-amazonaws Service rejected.

Full original title: PutTablePolicy Principal: nonexistent role ARN BadRequestException/400; sts assumed-role
session ARN HTTP 200; Service example.com BadRequestException/400; own :root HTTP 200
