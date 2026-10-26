<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-063: GetTableBucketPolicy re-serializes the policy: whitespace stripped, Version moved first, one-element Action array collapsed to a string
_Full entry and notes of one finding; its summary entry is in
[tablebucket-policy.md](../tablebucket-policy.md). Generated from aws-api-semantics `services/s3tables` (model
2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-063"></a>**S3T-TABLEBUCKET-063** `normalization` · impact high · handled · verified 2026-10-13
  **GetTableBucketPolicy re-serializes the policy: whitespace stripped, Version moved first, one-element Action array collapsed to a string**
  Put of a 2-space-indented document with trailing newline, Statement before Version, Sid and Action as a
  one-element list -> HTTP 200; Get returned a minified document with Version first and
  "Action":"s3tables:GetTable" (array collapsed to a scalar). Put of the same statement already minified
  (Action still a list) -> Get again collapsed the list to a scalar, so even a minified document is not echoed
  byte-for-byte. Put with Action as a scalar string -> Get returned the sent string byte-identical. Member
  order inside a statement was preserved (Sid, Effect, Principal, Action, Resource).
  - ACK: is_iam_policy, compare.is_ignored+delta_pre_compare · ops: PutTableBucketPolicy, GetTableBucketPolicy
    · fields: resourcePolicy
  - repro: Put pretty-printed policy -> Get -> compare strings; Put minified -> Get; Put with scalar Action ->
    Get
  - measurements: n=3
  - handling: handled via `generator.yaml:234-238; pkg/resource/table_bucket/delta.go:69-75`
  - related: [S3T-TABLEBUCKETPOLICY-017](../tablebucket-policy.md#s3t-tablebucketpolicy-017), [S3T-TABLEBUCKETPOLICY-004](../tablebucket-policy.md#s3t-tablebucketpolicy-004), [S3T-TABLEBUCKET-111](../tablebucket-policy.md#s3t-tablebucket-111),
    [S3T-TABLEBUCKETPOLICY-012](../tablebucket-policy.md#s3t-tablebucketpolicy-012), [S3T-TABLEBUCKETPOLICY-013](../tablebucket-policy.md#s3t-tablebucketpolicy-013), [S3T-TABLE-152](../table-policy.md#s3t-table-152), [S3T-TABLEBUCKETPOLICY-008](../tablebucket-policy.md#s3t-tablebucketpolicy-008),
    [S3T-TABLEPOLICY-005](../table-policy.md#s3t-tablepolicy-005), [S3T-TABLE-154](../table-policy.md#s3t-table-154), [S3T-TABLE-166](../table-policy.md#s3t-table-166), [S3T-TABLEPOLICY-003](../table-policy.md#s3t-tablepolicy-003), [S3T-TABLEPOLICY-008](../table-policy.md#s3t-tablepolicy-008) · hypotheses:
    H-S-015 · evidence: tablebucket/sub-resources/policy-get-put-delete

## Notes

H-S-015 confirmed (3 documents): compare policies semantically (parse, treat one-element arrays as equal to
scalars); a string comparison of spec vs Get differs unless the user writes the canonical form (minified,
Version first, scalar single values).

Contradiction with [S3T-TABLE-152](../table-policy.md#s3t-table-152): 063 states 'member order inside a statement was preserved (Sid, Effect,
Principal, Action, Resource)' but its input was already in that order; 152 sent Sid, Resource, Action, Effect,
Principal and read back Sid, Effect, Principal, Action, Resource - the normalizer re-keys statements to the
canonical IAM order (same engine at both levels per [S3T-TABLEPOLICY-003](../table-policy.md#s3t-tablepolicy-003)/-008) Resolution: keep both;
[S3T-TABLE-152](../table-policy.md#s3t-table-152) is canonical for statement key order; 063's canonical-order input could not reveal the re-keying
