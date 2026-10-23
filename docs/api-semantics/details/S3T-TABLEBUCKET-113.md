<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-113: Metrics configuration Put/Get/Delete as a role with s3tables:* Allow and cloudwatch:* Deny -> HTTP 204 / HTTP 200 / HTTP 204...
_Full entry and notes of one finding; its summary entry is in
[tablebucket-maintenance-metrics.md](../tablebucket-maintenance-metrics.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-113"></a>**S3T-TABLEBUCKET-113** `prerequisite` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Metrics configuration Put/Get/Delete as a role with s3tables:* Allow and cloudwatch:* Deny -> HTTP 204 / HTTP 200 / HTTP 204...**
  A fresh IAM role with an inline policy {Allow s3tables:* on *, Deny cloudwatch:* on *} (its own
  cloudwatch:ListMetrics -> AccessDenied 403 '... with an explicit deny in an identity-based policy'):
  PutTableBucketMetricsConfiguration -> HTTP 204 (attempts while IAM propagated: ['HTTP 204']),
  GetTableBucketMetricsConfiguration -> HTTP 200 {'tableBucketARN':
  'arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-5b1f89-cwd', 'id': 'EntireBucket'},
  DeleteTableBucketMetricsConfiguration -> HTTP 204, Get after Delete -> NotFoundException/404. iam:ListRoles
  PathPrefix=/aws-service-role/ before (14 roles) and 5 s after enabling metrics: no new role.
  - ACK: docs-only · ops: PutTableBucketMetricsConfiguration, GetTableBucketMetricsConfiguration,
    DeleteTableBucketMetricsConfiguration · fields: tableBucketARN
  - repro: CreateRole (trust: account root; inline Allow s3tables:*, Deny cloudwatch:*) -> AssumeRole ->
    Put/Get/Delete metrics -> ListRoles /aws-service-role/ diff
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-093](../tablebucket-maintenance-metrics.md#s3t-tablebucket-093), [S3T-TABLEBUCKET-094](../tablebucket-maintenance-metrics.md#s3t-tablebucket-094), [S3T-TABLEBUCKET-095](../tablebucket-maintenance-metrics.md#s3t-tablebucket-095), [S3T-TABLEBUCKET-099](../tablebucket-maintenance-metrics.md#s3t-tablebucket-099),
    [S3T-TABLEBUCKET-100](../tablebucket-maintenance-metrics.md#s3t-tablebucket-100) · hypotheses: H-S-095 · evidence: tablebucket/dependencies/iam-cloudwatch-deny-metrics

## Notes

H-S-095 confirmed: no CloudWatch permission and no service-linked role is needed to enable bucket metrics.
Single role, single bucket.

Full original title: Metrics configuration Put/Get/Delete as a role with s3tables:* Allow and cloudwatch:*
Deny -> HTTP 204 / HTTP 200 / HTTP 204; new service-linked roles: none
