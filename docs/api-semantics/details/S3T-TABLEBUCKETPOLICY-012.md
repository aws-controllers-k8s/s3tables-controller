<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKETPOLICY-012: GetTableBucketPolicy re-shapes 7 more forms (Statement object->array, 1-elem Resource/Principal/Condition lists->scalar...
_Full entry and notes of one finding; its summary entry is in
[tablebucket-policy.md](../tablebucket-policy.md). Generated from aws-api-semantics `services/s3tables` (model
2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucketpolicy-012"></a>**S3T-TABLEBUCKETPOLICY-012** `normalization` · impact medium · SUSPECTED CONTROLLER BUG · verified 2026-10-14
  **GetTableBucketPolicy re-shapes 7 more forms (Statement object->array, 1-elem Resource/Principal/Condition lists->scalar...**
  Put of 25 policy variants on one bucket, each followed by GetTableBucketPolicy and a structural compare.
  RE-SHAPED: Statement given as a single object -> 1-element array; one-element Resource list -> scalar
  string; one-element Principal.AWS list -> scalar; Principal.AWS list [root, root] -> de-duplicated scalar;
  Condition Bool value JSON false -> the string "false"; Condition StringEquals one-element list -> scalar;
  Effect "ALLOW" -> "Allow". RETURNED STRUCTURALLY IDENTICAL (only minified): action name case as sent
  (s3tables:gettablebucket, s3tables:LISTNAMESPACES), wildcard s3tables:Get*, two-element Resource list,
  Allow+NotAction, Allow+NotResource, two statements with the same Sid (accepted), two byte-identical
  statements (both kept), statements without Sid, Sid "", Sid caf\u00e9 (returned as the raw UTF-8 char),
  explicit Version 2008-10-17 (kept), top-level Id (kept). REJECTED at Put: S3Tables:GetTableBucket
  (capitalised service prefix) -> BadRequestException (HTTP 400) 'Invalid Action'; Principal as a bare ARN
  string -> BadRequestException (HTTP 400) 'The specified policy is malformed'; Condition operator 'bool' /
  key 'AWS:securetransport' -> BadRequestException (HTTP 400) 'The specified policy is malformed';
  Allow+NotPrincipal -> BadRequestException (HTTP 400) 'Public policies cannot be applied to S3 table buckets
  and tables.'. No read-back was byte-identical to the sent text.
  - ACK: is_iam_policy, compare.is_ignored+delta_pre_compare · ops: PutTableBucketPolicy, GetTableBucketPolicy
    · fields: resourcePolicy
  - repro: Put each variant (Statement object, action case, one-element lists, bare principal, JSON bool
    condition, Not* members, duplicate Sids, unicode, Version 2008, Id) -> Get -> compare
  - handling: suspected controller bug - see Handling gaps
  - related: [S3T-TABLEBUCKET-063](../tablebucket-policy.md#s3t-tablebucket-063), [S3T-TABLEBUCKET-111](../tablebucket-policy.md#s3t-tablebucket-111), [S3T-TABLE-152](../table-policy.md#s3t-table-152), [S3T-TABLE-154](../table-policy.md#s3t-table-154),
    [S3T-TABLEBUCKETPOLICY-013](../tablebucket-policy.md#s3t-tablebucketpolicy-013), [S3T-TABLEBUCKETPOLICY-004](../tablebucket-policy.md#s3t-tablebucketpolicy-004), [S3T-TABLEBUCKETPOLICY-017](../tablebucket-policy.md#s3t-tablebucketpolicy-017), [S3T-TABLEBUCKET-088](../tablebucket-policy.md#s3t-tablebucket-088),
    [S3T-TABLEBUCKET-089](../tablebucket-policy.md#s3t-tablebucket-089), [S3T-TABLEBUCKET-090](../tablebucket-policy.md#s3t-tablebucket-090), [S3T-TABLEBUCKETPOLICY-001](../tablebucket-policy.md#s3t-tablebucketpolicy-001), [S3T-TABLEBUCKETPOLICY-002](../tablebucket-policy.md#s3t-tablebucketpolicy-002),
    [S3T-TABLEBUCKETPOLICY-014](../tablebucket-policy.md#s3t-tablebucketpolicy-014), [S3T-TABLE-157](../table-policy.md#s3t-table-157), [S3T-TABLEPOLICY-008](../table-policy.md#s3t-tablepolicy-008), [S3T-TABLE-166](../table-policy.md#s3t-table-166), [S3T-TABLEPOLICY-003](../table-policy.md#s3t-tablepolicy-003),
    [S3T-TABLE-167](../table-policy.md#s3t-table-167), [S3T-TABLE-162](../table-policy.md#s3t-table-162), [S3T-TABLEPOLICY-009](../table-policy.md#s3t-tablepolicy-009) · evidence:
    tablebucketpolicy/creative/normalization-extras

## Notes

Extends the known re-serialization findings (minify, Version injection, Effect case, 1-element Action
collapse, account-id -> :root) with the remaining shapes a semantic comparator must treat as equal or a
controller will re-Put on every reconcile. Single run per cell; same engine at both levels: see the other
level's finding.

Full original title: GetTableBucketPolicy re-shapes 7 more forms (Statement object->array, 1-elem
Resource/Principal/Condition lists->scalar, JSON bool->string, dup principals, Effect case); 4 rejected

Suspected controller bug confirmed by evidence: not flagged; evidence-driven.
ackcompare.IAMPolicyDocumentEqual tolerates only whitespace, key order, statement order, Statement
object->array and one-element list collapsing; it reports a difference for a service-inserted Version (desired
omits Version, read-back has '2008-10-17'), for Effect case folding ('allow'/'ALLOW' -> 'Allow'), for
de-duplicated principals ([root, root] -> root), for Condition Bool false -> "false" and for account-id ->
:root, so a TableBucket whose policy uses any of these forms carries a permanent Spec.Policy delta and
re-issues PutTableBucketPolicy on every reconcile (resync and every spec edit) without ever converging -
normalize desired (default Version, canonical Effect case, de-dup, stringify Bool values, expand bare account
ids) before comparing, or compare against the read-back of the last Put. code:
pkg/resource/table_bucket/delta.go:69-75 compares Spec.Policy with ackcompare.IAMPolicyDocumentEqual; runtime
v0.64.0 pkg/compare/iam_policy.go:31-46 unmarshals into micahhausler/aws-iam-policy Policy and calls
Policy.Equal, which is strict on Version (policy_statement.go:38-39) and Effect (policy_statement.go:86),
compares string lists by sorted slice with length check (string_slice.go:95-120, so de-dup fails) and keeps
bool and string condition values in separate buckets (condition_value.go:56-60, so false != "false"); verified
by a unit test outside the repo feeding the sent/read-back pairs through ackcompare.IAMPolicyDocumentEqual at
runtime v0.64.0: version-inserted=false, effect-case-fold=false, duplicate-principals-removed=false,
bool-false-stringified=false, account-id-to-root=false, role-arn-to-aroa=false;
one-element-lists-collapsed=true, statement-object-to-array=true, whitespace-key-order-only=true.
pkg/resource/table_bucket/hooks.go:197-251 customUpdateTableBucket returns the desired spec untouched, so the
CR is never rewritten to the read-back and the delta recurs. The ARN -> AROA rewrite after a role deletion is
a legitimate delta, but the re-Put then hits the terminal BadRequestException of [GT-S3T-010](../service.md#gt-s3t-010) (controller hooks catalog entry). finding:
[S3T-TABLEBUCKETPOLICY-012](../tablebucket-policy.md#s3t-tablebucketpolicy-012) (Statement object->array, 1-elem lists->scalar, [root,root]->root, Bool
false->"false", Effect ALLOW->Allow, account id -> :root), [S3T-TABLEBUCKET-111](../tablebucket-policy.md#s3t-tablebucket-111) (Version 2008-10-17 inserted,
Effect allow->Allow), [S3T-TABLEBUCKETPOLICY-013](../tablebucket-policy.md#s3t-tablebucketpolicy-013) (deleted role ARN -> AROA id).
