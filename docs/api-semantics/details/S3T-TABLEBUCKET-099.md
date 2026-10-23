<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-099: Metrics Put/Get/Delete with a foreign account id in the bucket ARN -> ['ForbiddenException/403']...
_Full entry and notes of one finding; its summary entry is in
[tablebucket-maintenance-metrics.md](../tablebucket-maintenance-metrics.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-099"></a>**S3T-TABLEBUCKET-099** `error-code` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Metrics Put/Get/Delete with a foreign account id in the bucket ARN -> ['ForbiddenException/403']...**
  <ACCOUNT>/put -> ForbiddenException (HTTP 403) 'Unauthorized'; <ACCOUNT>/get -> ForbiddenException (HTTP 403)
  'Unauthorized'; <ACCOUNT>/delete -> ForbiddenException (HTTP 403) 'Unauthorized'; <ACCOUNT>/put ->
  ForbiddenException (HTTP 403) 'Unauthorized'; <ACCOUNT>/get -> ForbiddenException (HTTP 403) 'Unauthorized';
  <ACCOUNT>/delete -> ForbiddenException (HTTP 403) 'Unauthorized'. Contrast:
  PutTableBucketMetricsConfiguration on the caller's account with a never-existing bucket name ->
  NotFoundException (HTTP 404) 'The specified bucket does not exist.'.
  - ACK: terminal_codes · ops: PutTableBucketMetricsConfiguration, GetTableBucketMetricsConfiguration,
    DeleteTableBucketMetricsConfiguration · fields: tableBucketARN
  - repro: Put/Get/Delete metrics with arn:aws:s3tables:<region>:<ACCOUNT>:bucket/<existing-name> and
    ...:<ACCOUNT>:...; Put on own-account never-existing name
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-093](../tablebucket-maintenance-metrics.md#s3t-tablebucket-093), [S3T-TABLEBUCKET-094](../tablebucket-maintenance-metrics.md#s3t-tablebucket-094), [S3T-TABLEBUCKET-095](../tablebucket-maintenance-metrics.md#s3t-tablebucket-095), [S3T-TABLEBUCKET-100](../tablebucket-maintenance-metrics.md#s3t-tablebucket-100),
    [S3T-TABLEBUCKET-113](../tablebucket-maintenance-metrics.md#s3t-tablebucket-113) · hypotheses: H-S-090 · evidence: tablebucket/error-taxonomy/subresource-arn-variants

## Notes

H-S-090 confirmed: all six calls (two foreign account ids x Put/Get/Delete) -> ForbiddenException (HTTP 403)
'Unauthorized'; the own-account never-existing name -> NotFoundException 404. Whether a bucket of that name
exists in the foreign accounts is unknown.

Full original title: Metrics Put/Get/Delete with a foreign account id in the bucket ARN ->
['ForbiddenException/403']; own account + never-existing name -> NotFoundException/404
