<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKETPOLICY-016: Bucket policy Deny of s3tables:DeleteTableBucketPolicy for principal B -> B's Delete AccessDeniedException 403; A (not denied) deletes 204
_Full entry and notes of one finding; its summary entry is in
[tablebucket-policy.md](../tablebucket-policy.md). Generated from aws-api-semantics `services/s3tables` (model
2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucketpolicy-016"></a>**S3T-TABLEBUCKETPOLICY-016** `error-code` · impact medium · handled · verified 2026-10-14
  **Bucket policy Deny of s3tables:DeleteTableBucketPolicy for principal B -> B's Delete AccessDeniedException 403; A (not denied) deletes 204**
  A PutTableBucketPolicy(P2 = P1 + {Effect Deny, Principal AWS <B's role ARN>, Action
  s3tables:DeleteTableBucketPolicy, Resource <bucket ARN>}) -> HTTP 200. +3 s: B DeleteTableBucketPolicy ->
  AccessDeniedException (HTTP 403) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to
  perform: s3tables:DeleteTableBucketPolicy on resource:
  arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-4c97df-fu02 with an explicit deny in a resource-based
  policy'; second attempt -> AccessDeniedException (HTTP 403); B GetTableBucketPolicy -> HTTP 200; B
  GetTableBucket -> HTTP 200; A GetTableBucketPolicy -> HTTP 200; A DeleteTableBucketPolicy -> HTTP 204; then
  A Get -> NotFoundException (HTTP 404) and B Get -> NotFoundException (HTTP 404).
  - ACK: terminal_codes, custom_delete · ops: PutTableBucketPolicy, DeleteTableBucketPolicy,
    GetTableBucketPolicy · fields: resourcePolicy
  - repro: A: PutTableBucketPolicy with a Deny s3tables:DeleteTableBucketPolicy for B's role ARN -> B:
    DeleteTableBucketPolicy -> A: DeleteTableBucketPolicy
  - handling: handled via `generator.yaml:229-234; pkg/resource/table_bucket/hooks.go:325-332`
  - related: [S3T-TABLEBUCKET-045](../tablebucket-policy.md#s3t-tablebucket-045), [S3T-TABLEBUCKET-046](../tablebucket-policy.md#s3t-tablebucket-046), [S3T-TABLEBUCKET-129](../tablebucket-policy.md#s3t-tablebucket-129), [S3T-TABLEBUCKET-091](../tablebucket-policy.md#s3t-tablebucket-091),
    [S3T-TABLEBUCKETPOLICY-005](../tablebucket-policy.md#s3t-tablebucketpolicy-005), [S3T-TABLEBUCKETPOLICY-006](../tablebucket-policy.md#s3t-tablebucketpolicy-006), [S3T-TABLEBUCKETPOLICY-007](../tablebucket-policy.md#s3t-tablebucketpolicy-007), [S3T-TABLEPOLICY-009](../table-policy.md#s3t-tablepolicy-009) ·
    evidence: tablebucketpolicy/sub-resources/fu-02-delete-policy-principal-b

## Notes

FU-02 round 2. The denied principal gets the same wire code as an IAM denial (see behavior); a controller
whose role is denied by the policy it manages cannot remove that policy - and a Deny for '*' bricks the bucket
for everyone (the lab's pre-existing polnorm bucket is such a case; not touched here).
