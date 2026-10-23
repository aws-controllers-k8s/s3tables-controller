<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-152: GetTablePolicy returns the policy minified with statement keys reordered to Sid, Effect, Principal, Action, Resource
_Full entry and notes of one finding; its summary entry is in [table-policy.md](../table-policy.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-table-152"></a>**S3T-TABLE-152** `normalization` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **GetTablePolicy returns the policy minified with statement keys reordered to Sid, Effect, Principal, Action, Resource**
  Submitted a 2-statement policy pretty-printed (4-space indent, a tab, double spaces, trailing newline, 753
  bytes) with statement keys in the order Sid, Resource, Action, Effect, Principal. GetTablePolicy returned
  552 bytes: no whitespace at all, top-level order Version, Statement kept, every statement re-keyed in the
  order Sid, Effect, Principal, Action, Resource; statement order and all values unchanged (structurally
  equal). A string compare against the submitted document never matches.
  - ACK: is_document, compare.is_ignored+delta_pre_compare · ops: PutTablePolicy, GetTablePolicy · fields:
    resourcePolicy
  - repro: PutTablePolicy(pretty-printed 2-statement doc) -> GetTablePolicy -> compare bytes
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-063](../tablebucket-policy.md#s3t-tablebucket-063), [S3T-TABLEBUCKETPOLICY-008](../tablebucket-policy.md#s3t-tablebucketpolicy-008), [S3T-TABLEBUCKET-111](../tablebucket-policy.md#s3t-tablebucket-111), [S3T-TABLEBUCKETPOLICY-004](../tablebucket-policy.md#s3t-tablebucketpolicy-004),
    [S3T-TABLEPOLICY-005](../table-policy.md#s3t-tablepolicy-005), [S3T-TABLE-154](../table-policy.md#s3t-table-154), [S3T-TABLE-166](../table-policy.md#s3t-table-166), [S3T-TABLEPOLICY-003](../table-policy.md#s3t-tablepolicy-003), [S3T-TABLEPOLICY-008](../table-policy.md#s3t-tablepolicy-008),
    [S3T-TABLEBUCKETPOLICY-012](../tablebucket-policy.md#s3t-tablebucketpolicy-012) · hypotheses: H-U-005, H-U-006 · evidence: table/sub-resources/policy-lifecycle

## Notes

H-U-005 qualified: minified and structurally equal, but key order is NOT preserved (canonical IAM order);
H-U-006 refuted (not byte-for-byte). Single document, n=1 but deterministic across the 2 identical Puts.

Contradiction with [S3T-TABLEBUCKET-063](../tablebucket-policy.md#s3t-tablebucket-063): 063 states 'member order inside a statement was preserved (Sid,
Effect, Principal, Action, Resource)' but its input was already in that order; 152 sent Sid, Resource, Action,
Effect, Principal and read back Sid, Effect, Principal, Action, Resource - the normalizer re-keys statements
to the canonical IAM order (same engine at both levels per [S3T-TABLEPOLICY-003](../table-policy.md#s3t-tablepolicy-003)/-008) Resolution: keep both;
[S3T-TABLE-152](../table-policy.md#s3t-table-152) is canonical for statement key order; 063's canonical-order input could not reveal the re-keying
