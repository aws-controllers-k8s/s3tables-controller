<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEPOLICY-009: Policy Deny binds for Resource = uuid table ARN, <bucket>/table/*, <bucket>/*, '*'...
_Full entry and notes of one finding; its summary entry is in [table-policy.md](../table-policy.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-tablepolicy-009"></a>**S3T-TABLEPOLICY-009** `request-validation` · impact high · unhandled (not handled in controller) · verified 2026-10-15, re-verified
  **Policy Deny binds for Resource = uuid table ARN, <bucket>/table/*, <bucket>/*, '*'...**
  Deny(Principal '*', Action s3tables:GetTableMaintenanceConfiguration, Resource=<form>) then the denied call
  on the table by the caller itself (1 s after Put). Table policy: exact_table_arn -> Put HTTP 200, call
  AccessDeniedException/403; name_based_pseudo_arn -> Put HTTP 200, call HTTP 200; bucket_arn -> Put HTTP 200,
  call HTTP 200; bucket_table_wildcard -> Put HTTP 200, call AccessDeniedException/403; bucket_slash_wildcard
  -> Put HTTP 200, call AccessDeniedException/403; star -> Put HTTP 200, call AccessDeniedException/403;
  other_table_arn -> Put HTTP 200, call HTTP 200; exact_table_arn_list_with_other -> Put HTTP 200, call
  AccessDeniedException/403. Bucket policy with the same statements: exact_table_arn -> Put HTTP 200, call
  AccessDeniedException/403; bucket_table_wildcard -> Put HTTP 200, call AccessDeniedException/403;
  name_based_pseudo_arn -> Put HTTP 200, call HTTP 200; bucket_arn -> Put HTTP 200, call HTTP 200;
  bucket_slash_wildcard -> Put HTTP 200, call AccessDeniedException/403. A table-policy Deny never affected
  the sibling table (all HTTP 200). Every Put returned ['HTTP 200']; DeleteTablePolicy/DeleteTableBucketPolicy
  restored HTTP 200 on the next call in all cases.
  - ACK: is_iam_policy, docs-only, compare.is_ignored+delta_pre_compare · ops: PutTablePolicy,
    PutTableBucketPolicy, GetTableMaintenanceConfiguration · fields: resourcePolicy, Resource
  - repro: PutTablePolicy(Deny * GetTableMaintenanceConfiguration Resource=<form>) ->
    GetTableMaintenanceConfiguration -> DeleteTablePolicy, per form; same via PutTableBucketPolicy
  - measurements: table_policy_binding_forms=[exact_table_arn, bucket_table_wildcard, bucket_slash_wildcard,
    star, exact_table_arn_list_with_other], table_policy_inert_forms=[name_based_pseudo_arn, bucket_arn,
    other_table_arn], table_policy_rejected_forms=[], bucket_policy_binding_forms=[exact_table_arn,
    bucket_table_wildcard, bucket_slash_wildcard], bucket_policy_inert_forms=[name_based_pseudo_arn,
    bucket_arn], bucket_policy_rejected_forms=[], n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-163](../table-policy.md#s3t-table-163), [S3T-TABLE-167](../table-policy.md#s3t-table-167), [S3T-TABLE-213](../table-maintenance.md#s3t-table-213), [S3T-TABLE-251](../table-policy.md#s3t-table-251), [S3T-TABLEBUCKET-091](../tablebucket-policy.md#s3t-tablebucket-091),
    [S3T-TABLEBUCKETPOLICY-005](../tablebucket-policy.md#s3t-tablebucketpolicy-005), [S3T-TABLEBUCKETPOLICY-016](../tablebucket-policy.md#s3t-tablebucketpolicy-016), [S3T-TABLEBUCKETPOLICY-007](../tablebucket-policy.md#s3t-tablebucketpolicy-007), [S3T-TABLEBUCKET-088](../tablebucket-policy.md#s3t-tablebucket-088),
    [S3T-TABLE-162](../table-policy.md#s3t-table-162), [S3T-TABLEBUCKETPOLICY-012](../tablebucket-policy.md#s3t-tablebucketpolicy-012), [S3T-TABLE-004](../table-policy.md#s3t-table-004) · evidence: table/creative/policy-resource-binding

## Notes

extends [S3T-TABLE-163](../table-policy.md#s3t-table-163)/-167 (any Resource accepted) with which forms are actually enforced. Decision rule: a
controller that lets users write Resource with the name-based pseudo ARN or the bucket ARN in a TablePolicy
produces a policy that reads back fine but protects nothing; the controller should substitute the table's real
ARN (uuid form, from status) and warn on the other forms. Single pass per form.

Full original title: Policy Deny binds for Resource = uuid table ARN, <bucket>/table/*, <bucket>/*, '*';
name-based <bucket>/table/<ns>/<name> and bare bucket ARN are inert
