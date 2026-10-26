<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEPOLICY-002: PutTablePolicy right after CreateRole: BadRequestException 400 for 8.57-8.66 s (n=3) then 200 - a transient 400 a controller must retry
_Full entry and notes of one finding; its summary entry is in [table-policy.md](../table-policy.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-tablepolicy-002"></a>**S3T-TABLEPOLICY-002** `error-code` · impact high · handled · verified 2026-10-14
  **PutTablePolicy right after CreateRole: BadRequestException 400 for 8.57-8.66 s (n=3) then 200 - a transient 400 a controller must retry**
  CreateRole -> PutTablePolicy(Principal=<new role ARN>) retried every 1 s: rejected [8, 8, 8] times (per
  trial) with ['BadRequestException/400'] ["Invalid principal in policy"] before the first HTTP 200 at [8.66,
  8.63, 8.57] s. The same code and message are returned for a role that never existed, so the 400 is not
  distinguishable from a permanent spec error by code or text.
  - ACK: requeue, terminal_codes, references · ops: PutTablePolicy, PutTableBucketPolicy · fields:
    resourcePolicy, Principal
  - repro: iam CreateRole -> s3tables PutTablePolicy with Principal.AWS=<role ARN> at +0 ms, retry 1/s until
    200 (x3 roles)
  - measurements: ok_at_s=[8.66, 8.63, 8.57], n_rejected=[8, 8, 8], n=3
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLE-164](../table-policy.md#s3t-table-164), [S3T-TABLEREPLICATION-009](../table-replication.md#s3t-tablereplication-009), [S3T-TABLEBUCKETPOLICY-001](../tablebucket-policy.md#s3t-tablebucketpolicy-001), [S3T-TABLEBUCKETPOLICY-013](../tablebucket-policy.md#s3t-tablebucketpolicy-013),
    [S3T-TABLE-165](../table-policy.md#s3t-table-165), [S3T-TABLEPOLICY-004](../table-policy.md#s3t-tablepolicy-004), [S3T-TABLE-157](../table-policy.md#s3t-table-157) · evidence: table/creative/policy-principal-lifecycle

## Notes

extends [S3T-TABLE-164](../table-policy.md#s3t-table-164) (nonexistent role -> 400) with the propagation window after CreateRole measured from the
controller's call sequence. Decision rule: treat BadRequestException 'invalid principal' on Put*Policy as
retryable for at least ~30 s before surfacing it as terminal; a Role reference should be resolved and the
policy requeued, not marked failed, on this message.
