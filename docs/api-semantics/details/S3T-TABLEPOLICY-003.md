<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEPOLICY-003: GetTablePolicy is a fixed point: Put(observed) reads back byte-identical (table and bucket); desired minified != observed
_Full entry and notes of one finding; its summary entry is in [table-policy.md](../table-policy.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-tablepolicy-003"></a>**S3T-TABLEPOLICY-003** `normalization` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **GetTablePolicy is a fixed point: Put(observed) reads back byte-identical (table and bucket); desired minified != observed**
  Two-statement document (Effect 'allow' lower-case, 1-elem Principal list, 2-elem Action list, 1-elem
  Resource list, two Condition operators, Deny Principal '*'): Put -> HTTP 200; Get -> normalized form;
  Put(that form) -> HTTP 200; Get again byte-identical: True. Same through
  PutTableBucketPolicy/GetTableBucketPolicy: byte-identical True, same normalizer as the table path: True.
  Rules seen: top-level keys ['Version', 'Statement'], statement keys ['Sid', 'Effect', 'Principal', 'Action',
  'Resource', 'Condition'], Effect ['Allow', 'Deny'], principal list collapsed True, 2-elem Action kept as
  list True, 1-elem Resource collapsed True, Condition keys ['StringEquals', 'Bool'], statement order kept
  True. json.loads(observed) == desired: False.
  - ACK: is_iam_policy, compare.is_ignored+delta_pre_compare · ops: PutTablePolicy, GetTablePolicy,
    PutTableBucketPolicy, GetTableBucketPolicy · fields: resourcePolicy
  - repro: PutTablePolicy(desired) -> GetTablePolicy -> PutTablePolicy(observed) -> GetTablePolicy; same via
    the bucket policy ops
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-152](../table-policy.md#s3t-table-152), [S3T-TABLE-154](../table-policy.md#s3t-table-154), [S3T-TABLEBUCKET-063](../tablebucket-policy.md#s3t-tablebucket-063), [S3T-TABLEBUCKET-111](../tablebucket-policy.md#s3t-tablebucket-111), [S3T-TABLE-166](../table-policy.md#s3t-table-166),
    [S3T-TABLEPOLICY-008](../table-policy.md#s3t-tablepolicy-008), [S3T-TABLEBUCKETPOLICY-012](../tablebucket-policy.md#s3t-tablebucketpolicy-012) · evidence: table/creative/policy-principal-lifecycle

## Notes

extends [S3T-TABLE-152](../table-policy.md#s3t-table-152)/-154 and [S3T-TABLEBUCKET-063](../tablebucket-policy.md#s3t-tablebucket-063)/-111 with the controller-relevant property: the
normalization is idempotent, so a controller can store normalize(desired)=Get-after-first-Put and compare
strings, or compare semantically (lists vs scalars and Effect case differ); a raw string compare of spec vs
Get loops forever. Also records Condition/multi-statement handling. Single run.
