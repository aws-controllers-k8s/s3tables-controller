<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEPOLICY-008: GetTablePolicy re-shapes 7 more forms (Statement object->array, 1-elem Resource/Principal/Condition lists->scalar, JSON bool->string...
_Full entry and notes of one finding; its summary entry is in [table-policy.md](../table-policy.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-tablepolicy-008"></a>**S3T-TABLEPOLICY-008** `normalization` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **GetTablePolicy re-shapes 7 more forms (Statement object->array, 1-elem Resource/Principal/Condition lists->scalar, JSON bool->string...**
  Put of 25 policy variants on one table, each followed by GetTablePolicy and a structural compare. RE-SHAPED:
  Statement given as a single object -> 1-element array; one-element Resource list -> scalar string;
  one-element Principal.AWS list -> scalar; Principal.AWS list [root, root] -> de-duplicated scalar; Condition
  Bool value JSON false -> the string "false"; Condition StringEquals one-element list -> scalar; Effect
  "ALLOW" -> "Allow". RETURNED STRUCTURALLY IDENTICAL (only minified): action name case as sent
  (s3tables:gettablebucket, s3tables:LISTNAMESPACES), wildcard s3tables:Get*, two-element Resource list,
  Allow+NotAction, Allow+NotResource, two statements with the same Sid (accepted), two byte-identical
  statements (both kept), statements without Sid, Sid "", Sid caf\u00e9 (returned as the raw UTF-8 char),
  explicit Version 2008-10-17 (kept), top-level Id (kept). REJECTED at Put: S3Tables:GetTableBucket
  (capitalised service prefix) -> BadRequestException (HTTP 400) 'Invalid Action'; Principal as a bare ARN
  string -> BadRequestException (HTTP 400) 'The specified policy is malformed'; Condition operator 'bool' /
  key 'AWS:securetransport' -> BadRequestException (HTTP 400) 'The specified policy is malformed';
  Allow+NotPrincipal -> BadRequestException (HTTP 400) 'Public policies cannot be applied to S3 table buckets
  and tables.'. No read-back was byte-identical to the sent text.
  - ACK: is_iam_policy, compare.is_ignored+delta_pre_compare · ops: PutTablePolicy, GetTablePolicy · fields:
    resourcePolicy
  - repro: Put each variant (Statement object, action case, one-element lists, bare principal, JSON bool
    condition, Not* members, duplicate Sids, unicode, Version 2008, Id) -> Get -> compare
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-063](../tablebucket-policy.md#s3t-tablebucket-063), [S3T-TABLEBUCKET-111](../tablebucket-policy.md#s3t-tablebucket-111), [S3T-TABLE-152](../table-policy.md#s3t-table-152), [S3T-TABLE-154](../table-policy.md#s3t-table-154), [S3T-TABLEBUCKET-089](../tablebucket-policy.md#s3t-tablebucket-089),
    [S3T-TABLE-157](../table-policy.md#s3t-table-157), [S3T-TABLEBUCKETPOLICY-012](../tablebucket-policy.md#s3t-tablebucketpolicy-012), [S3T-TABLE-166](../table-policy.md#s3t-table-166), [S3T-TABLEPOLICY-003](../table-policy.md#s3t-tablepolicy-003) · evidence:
    tablebucketpolicy/creative/normalization-extras

## Notes

Extends the known re-serialization findings (minify, Version injection, Effect case, 1-element Action
collapse, account-id -> :root) with the remaining shapes a semantic comparator must treat as equal or a
controller will re-Put on every reconcile. Single run per cell; same engine at both levels: yes.

Full original title: GetTablePolicy re-shapes 7 more forms (Statement object->array, 1-elem
Resource/Principal/Condition lists->scalar, JSON bool->string, dup principals, Effect case); 4 rejected
