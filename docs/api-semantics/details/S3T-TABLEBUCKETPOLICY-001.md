<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKETPOLICY-001: PutTableBucketPolicy with a nonexistent role principal -> BadRequestException 400...
_Full entry and notes of one finding; its summary entry is in
[tablebucket-policy.md](../tablebucket-policy.md). Generated from aws-api-semantics `services/s3tables` (model
2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucketpolicy-001"></a>**S3T-TABLEBUCKETPOLICY-001** `prerequisite` · impact medium · SUSPECTED CONTROLLER BUG · verified 2026-10-13
  **PutTableBucketPolicy with a nonexistent role principal -> BadRequestException 400...**
  PutTableBucketPolicy whose Principal.AWS is a role ARN that does not exist -> BadRequestException (HTTP 400)
  'Invalid principal in policy'. With the shard role created moments earlier: attempts [+4.9s
  BadRequestException/400, +7.2s 200]; first 200 at +7.2s after CreateRole.
  - ACK: requeue, is_iam_policy · ops: PutTableBucketPolicy · fields: resourcePolicy
  - repro: PutTableBucketPolicy(Principal.AWS=arn of a missing role) -> record; CreateRole -> immediately
    PutTableBucketPolicy(Principal=new role), retry every 2 s
  - measurements: first_policy_ok_since_role_create_s=[7.2], rejected_attempts_before_ok=[1],
    rejected_codes=[BadRequestException/400], n=1
  - handling: suspected controller bug - see Handling gaps
  - related: [S3T-TABLEBUCKET-088](../tablebucket-policy.md#s3t-tablebucket-088), [S3T-TABLEBUCKET-089](../tablebucket-policy.md#s3t-tablebucket-089), [S3T-TABLEBUCKET-090](../tablebucket-policy.md#s3t-tablebucket-090), [S3T-TABLEBUCKETPOLICY-002](../tablebucket-policy.md#s3t-tablebucketpolicy-002),
    [S3T-TABLEBUCKETPOLICY-014](../tablebucket-policy.md#s3t-tablebucketpolicy-014), [S3T-TABLEBUCKETPOLICY-012](../tablebucket-policy.md#s3t-tablebucketpolicy-012), [S3T-SERVICE-010](../service.md#s3t-service-010), [S3T-SERVICE-025](../service.md#s3t-service-025), [S3T-SERVICE-012](../service.md#s3t-service-012),
    [S3T-TABLE-164](../table-policy.md#s3t-table-164), [S3T-TABLEBUCKETPOLICY-013](../tablebucket-policy.md#s3t-tablebucketpolicy-013), [S3T-TABLEPOLICY-002](../table-policy.md#s3t-tablepolicy-002), [S3T-TABLE-165](../table-policy.md#s3t-table-165), [S3T-TABLEPOLICY-004](../table-policy.md#s3t-tablepolicy-004),
    [S3T-TABLE-157](../table-policy.md#s3t-table-157) · hypotheses: H-S-023 · evidence: tablebucket/dependencies/repl-setup

## Notes

H-S-023 confirmed (400 for a missing principal, transient 4xx right after CreateRole) (n=1 role; the policy
loop ran interleaved with the replication Put loop).

Full original title: PutTableBucketPolicy with a nonexistent role principal -> BadRequestException 400; with a
role created 7s earlier -> 200 after 1 rejected attempt(s)

Suspected controller bug confirmed by evidence: not flagged; evidence-driven. BadRequestException is terminal
for TableBucket, so the transient 'Invalid principal in policy' that PutTableBucketPolicy returns for ~7-9 s
after a referenced IAM role is created (or re-created) sets ACK.Terminal=True and the bucket is never
reconciled again until the spec changes - the controller has the 'strands on transient' sign, not the 'loops
on malformed document' sign; classify PutTableBucketPolicy 400s by message (retry 'Invalid principal in
policy' with a requeue, keep 'malformed' / 'Invalid Action' / 'no principals' / 'Public policies' terminal).
code: generator.yaml:267-268 TableBucket terminal_codes = [BadRequestException] -> pkg/resource/table_bucket/
sdk.go:400-402 terminalAWSError returns true for every BadRequestException; pkg/resource/table_bucket/
hooks.go:334-342 syncPolicy returns the PutTableBucketPolicy error unchanged with no message inspection;
manager.go:397-417 onError turns it into ackerr.Terminal; the create-time Synced=False nudge
(templates/hooks/table_bucket/sdk_create_post_set_output.go.tpl:8-11) makes the first Put land seconds after
CreateTableBucket, squarely inside the propagation window; runtime v0.64.0 reconciler.go:118 uses
GenerationChangedPredicate and reconciler.go:1701 returns Result{} on ackerr.Terminal, so no retry ever
happens. finding: [S3T-TABLEBUCKETPOLICY-001](../tablebucket-policy.md#s3t-tablebucketpolicy-001) (BadRequestException 400 'Invalid principal in policy' at +4.9 s,
200 at +7.2 s after CreateRole), [S3T-TABLEBUCKETPOLICY-013](../tablebucket-policy.md#s3t-tablebucketpolicy-013) (8 x 400 over 0-7.5 s, 200 at +8.6 s; same-name
re-create), [S3T-TABLEBUCKET-088](../tablebucket-policy.md#s3t-tablebucket-088) / [S3T-TABLEBUCKET-089](../tablebucket-policy.md#s3t-tablebucket-089) (the permanent cases share the same code).
