<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-089: PutTableBucketPolicy with Principal '*' and no Condition -> BadRequestException 400; {'AWS':'*'} -> BadRequestException 400...
_Full entry and notes of one finding; its summary entry is in
[tablebucket-policy.md](../tablebucket-policy.md). Generated from aws-api-semantics `services/s3tables` (model
2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-089"></a>**S3T-TABLEBUCKET-089** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **PutTableBucketPolicy with Principal '*' and no Condition -> BadRequestException 400; {'AWS':'*'} -> BadRequestException 400...**
  Allow s3tables:GetTable on <bucket>/table/* to Principal '*' -> BadRequestException (HTTP 400) ('Public
  policies cannot be applied to S3 table buckets and tables.'); Principal {'AWS':'*'} -> BadRequestException
  (HTTP 400) ('Public policies cannot be applied to S3 table buckets and tables.'); Principal '*' with
  Condition StringEquals aws:PrincipalAccount=<own account> -> HTTP 200. Baseline policy unchanged after the
  rejected Puts: True.
  - ACK: terminal_codes, docs-only · ops: PutTableBucketPolicy · fields: resourcePolicy
  - repro: Put policy with Principal '*' (no Condition); with {'AWS':'*'}; with '*' + aws:PrincipalAccount
    condition; Get after each
  - measurements: n=3
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-088](../tablebucket-policy.md#s3t-tablebucket-088), [S3T-TABLEBUCKET-090](../tablebucket-policy.md#s3t-tablebucket-090), [S3T-TABLEBUCKETPOLICY-001](../tablebucket-policy.md#s3t-tablebucketpolicy-001), [S3T-TABLEBUCKETPOLICY-002](../tablebucket-policy.md#s3t-tablebucketpolicy-002),
    [S3T-TABLEBUCKETPOLICY-014](../tablebucket-policy.md#s3t-tablebucketpolicy-014), [S3T-TABLEBUCKETPOLICY-012](../tablebucket-policy.md#s3t-tablebucketpolicy-012), [S3T-TABLE-157](../table-policy.md#s3t-table-157), [S3T-TABLEPOLICY-008](../table-policy.md#s3t-tablepolicy-008) · hypotheses:
    H-S-017 · evidence: tablebucket/weird-inputs/policy-validation-lockout

## Notes

H-S-017 confirmed: public-principal policy rejected with BadRequestException; with the account condition it
returned HTTP 200.

Full original title: PutTableBucketPolicy with Principal '*' and no Condition -> BadRequestException 400;
{'AWS':'*'} -> BadRequestException 400; '*' + aws:PrincipalAccount condition -> HTTP 200
