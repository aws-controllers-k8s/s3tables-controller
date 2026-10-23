<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-157: PutTablePolicy rejects public policies: Allow Principal "*" no Condition -> BadRequestException (HTTP 400); Deny "*" -> HTTP 200...
_Full entry and notes of one finding; its summary entry is in [table-policy.md](../table-policy.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-table-157"></a>**S3T-TABLE-157** `request-validation` · impact high · handled · verified 2026-10-13
  **PutTablePolicy rejects public policies: Allow Principal "*" no Condition -> BadRequestException (HTTP 400); Deny "*" -> HTTP 200...**
  Allow s3tables:GetTable Principal "*" no Condition -> BadRequestException (HTTP 400) 'Public policies cannot
  be applied to S3 table buckets and tables.'. Principal {"AWS":"*"} -> BadRequestException (HTTP 400) 'Public
  policies cannot be applied to S3 table buckets and tables.'. Deny GetTableData Principal "*" -> HTTP 200.
  Allow "*" + Condition ArnLike aws:PrincipalArn=<own role> -> HTTP 200. Allow "*" + IpAddress aws:SourceIp ->
  BadRequestException (HTTP 400) 'Public policies cannot be applied to S3 table buckets and tables.'. Allow to
  another account's :root -> BadRequestException (HTTP 400) 'Invalid principal in policy'. Service principal
  s3.amazonaws.com -> HTTP 200.
  - ACK: terminal_codes · ops: PutTablePolicy · fields: resourcePolicy
  - repro: PutTablePolicy with each Principal "*" variant on the fixture table
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLEBUCKET-089](../tablebucket-policy.md#s3t-tablebucket-089), [S3T-TABLEBUCKETPOLICY-012](../tablebucket-policy.md#s3t-tablebucketpolicy-012), [S3T-TABLEPOLICY-008](../table-policy.md#s3t-tablepolicy-008), [S3T-TABLEBUCKETPOLICY-001](../tablebucket-policy.md#s3t-tablebucketpolicy-001),
    [S3T-TABLEBUCKETPOLICY-013](../tablebucket-policy.md#s3t-tablebucketpolicy-013), [S3T-TABLEPOLICY-002](../table-policy.md#s3t-tablepolicy-002), [S3T-TABLE-165](../table-policy.md#s3t-table-165), [S3T-TABLEPOLICY-004](../table-policy.md#s3t-tablepolicy-004), [S3T-TABLE-164](../table-policy.md#s3t-table-164),
    [S3T-TABLE-162](../table-policy.md#s3t-table-162), [S3T-TABLE-166](../table-policy.md#s3t-table-166), [S3T-TABLE-167](../table-policy.md#s3t-table-167), [S3T-TABLEPOLICY-001](../table-policy.md#s3t-tablepolicy-001) · hypotheses: H-U-026 · evidence:
    table/sub-resources/policy-lifecycle

## Notes

H-U-026 refuted: there IS a public-policy guard (BadRequestException 400 'Public policies cannot be applied to
S3 table buckets and tables.').

Full original title: PutTablePolicy rejects public policies: Allow Principal "*" no Condition ->
BadRequestException (HTTP 400); Deny "*" -> HTTP 200; "*"+aws:PrincipalArn -> HTTP 200
