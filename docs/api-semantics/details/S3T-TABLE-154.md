<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-154: GetTablePolicy: account-id principal rewritten to :root ARN, 1-elem principal list collapsed, 1-elem Action list collapsed
_Full entry and notes of one finding; its summary entry is in [table-policy.md](../table-policy.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-table-154"></a>**S3T-TABLE-154** `normalization` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **GetTablePolicy: account-id principal rewritten to :root ARN, 1-elem principal list collapsed, 1-elem Action list collapsed**
  PutTablePolicy with Principal {"AWS":"<account-id>"} read back as {"AWS": "arn:aws:iam::<ACCOUNT>:root"};
  {"AWS":["<role ARN>"]} read back as {"AWS": "arn:aws:iam::<ACCOUNT>:role/<PRINCIPAL>"}; "*" (with
  aws:PrincipalAccount Condition) read back as '*'; {"AWS":"*"} read back as {'AWS': '*'}; Action
  ["s3tables:GetTable"] read back as 's3tables:GetTable'. Statement given as a single object (not array) ->
  Put HTTP 200, read back Statement type list.
  - ACK: is_iam_policy, compare.is_ignored+delta_pre_compare · ops: PutTablePolicy, GetTablePolicy · fields:
    resourcePolicy
  - repro: PutTablePolicy with account-id, [role ARN], "*"+Condition, {AWS:"*"}+Condition principals and a
    1-element Action list -> GetTablePolicy
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETPOLICY-008](../tablebucket-policy.md#s3t-tablebucketpolicy-008), [S3T-TABLEBUCKET-063](../tablebucket-policy.md#s3t-tablebucket-063), [S3T-TABLEBUCKET-111](../tablebucket-policy.md#s3t-tablebucket-111), [S3T-TABLEBUCKETPOLICY-004](../tablebucket-policy.md#s3t-tablebucketpolicy-004),
    [S3T-TABLEPOLICY-005](../table-policy.md#s3t-tablepolicy-005), [S3T-TABLE-152](../table-policy.md#s3t-table-152), [S3T-TABLE-166](../table-policy.md#s3t-table-166), [S3T-TABLEPOLICY-003](../table-policy.md#s3t-tablepolicy-003), [S3T-TABLEPOLICY-008](../table-policy.md#s3t-tablepolicy-008),
    [S3T-TABLEBUCKETPOLICY-012](../tablebucket-policy.md#s3t-tablebucketpolicy-012) · hypotheses: H-U-007 · evidence: table/sub-resources/policy-lifecycle

## Notes

H-U-007 confirmed: acct->root True, list collapse True, '*' kept True. Unconditioned '*' could not be tested
(rejected as public, see H-U-026 finding). The unconditioned '*' is rejected as public, so the '*' round trip
was observed with an aws:PrincipalAccount Condition: returned as the bare string '*' and {'AWS':'*'} kept as
object - no canonicalization between the two forms.
