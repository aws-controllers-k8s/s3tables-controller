<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-110: PutTableBucketPolicy accepts unknown s3tables:* actions, Resource '*', s3/foreign ARNs; rejects Action '*' and a nonexistent principal
_Full entry and notes of one finding; its summary entry is in
[tablebucket-policy.md](../tablebucket-policy.md). Generated from aws-api-semantics `services/s3tables` (model
2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-110"></a>**S3T-TABLEBUCKET-110** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **PutTableBucketPolicy accepts unknown s3tables:* actions, Resource '*', s3/foreign ARNs; rejects Action '*' and a nonexistent principal**
  Accepted with HTTP 200 and stored as sent (17 single-statement documents on one bucket): Action
  's3tables:*', 's3tables:FooBarBaz', 's3tables:gettable' (case kept), 's3tables:Get*',
  ['s3tables:GetTable','s3tables:NotReal']; Resource '*', the bucket ARN itself, a bucket ARN in another
  region, in another account, 'arn:aws:s3:::some-bucket/*', a nonexistent same-account bucket; Principal
  {'Service':'maintenance.s3tables.amazonaws.com'}; Version '2008-10-17'. Rejected: Action '*' ->
  BadRequestException 400 'Invalid Action'; Principal = ARN of a role that does not exist ->
  BadRequestException 400 'Invalid principal in policy'.
  - ACK: is_iam_policy, docs-only · ops: PutTableBucketPolicy, GetTableBucketPolicy · fields: resourcePolicy
  - repro: Put each document on a fresh bucket, Get, Delete; record code/HTTP/message
  - measurements: n=17
  - handling: not handled in the controller (as of commit 1eda357)
  - hypotheses: H-S-016 · evidence: tablebucket/creative/policy-validation-boundary

## Notes

Creative follow-up to H-S-016: validation is syntactic plus a principal-existence check and an
action-namespace check ('s3tables:' prefix required; '*' and other services' actions are 'Invalid Action');
action names after the prefix and the Resource ARN are not validated against anything. A controller cannot
rely on Put to catch typos in actions or a Resource pointing at another bucket.

Full original title: PutTableBucketPolicy accepts unknown s3tables:* actions, Resource '*', s3 ARNs and
other-account/region bucket ARNs; rejects Action '*' and a nonexistent role principal
