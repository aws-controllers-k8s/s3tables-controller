<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-100: Metrics Put ARN variants: upper-case, trailing slash, table ARN, empty region, s3, other region, aws-cn, bare name -> 400; underscore -> 404
_Full entry and notes of one finding; its summary entry is in
[tablebucket-maintenance-metrics.md](../tablebucket-maintenance-metrics.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-100"></a>**S3T-TABLEBUCKET-100** `identity` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Metrics Put ARN variants: upper-case, trailing slash, table ARN, empty region, s3, other region, aws-cn, bare name -> 400; underscore -> 404**
  On an existing bucket with one table: uppercase-name
  (aws:s3tables:us-west-2:<ACCOUNT>:bucket/ACKQ-28A1D7-EARN) -> BadRequestException (HTTP 400) 'The specified
  bucket name isn't valid. Specify a different bucket name, and then try again.'; underscore-name
  (aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq_28a1d7_earn) -> NotFoundException (HTTP 404) 'The specified
  bucket does not exist.'; trailing-slash (aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-28a1d7-earn/) ->
  BadRequestException (HTTP 400) 'The specified bucket name isn't valid. Specify a different bucket name, and
  then try again.'; table-arn (aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-28a1d7-earn/) ->
  BadRequestException (HTTP 400) 'The specified bucket name isn't valid. Specify a different bucket name, and
  then try again.'; empty-region (aws:s3tables::<ACCOUNT>:bucket/ackq-28a1d7-earn) -> BadRequestException
  (HTTP 400) 'The specified table bucket ARN is not valid.'; service-s3
  (aws:s3:us-west-2:<ACCOUNT>:bucket/ackq-28a1d7-earn) -> BadRequestException (HTTP 400) 'The specified table
  bucket ARN is not valid.'; other-region (aws:s3tables:us-east-1:<ACCOUNT>:bucket/ackq-28a1d7-earn) ->
  BadRequestException (HTTP 400) 'The provided Region does not match the Region of the resource.';
  partition-aws-cn (aws-cn:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-28a1d7-ea) -> BadRequestException (HTTP 400)
  'The provided Partition does not match the Partition of the resource.'; name-only (ackq-28a1d7-earn) ->
  BadRequestException (HTTP 400) 'The specified table bucket ARN is not valid.'.
  GetTableBucketMetricsConfiguration with the canonical ARN afterwards -> NotFoundException (HTTP 404) 'No
  such metric configuration exists' (no configuration was created by any variant).
  - ACK: is_arn_primary_key, terminal_codes · ops: PutTableBucketMetricsConfiguration,
    GetTableBucketMetricsConfiguration · fields: tableBucketARN
  - repro: Put metrics with each ARN variant (upper-case name, underscore name, trailing slash, table ARN,
    empty region, service s3, other region, aws-cn partition, bare name) -> canonical Get
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-001](../tablebucket-encryption-storage.md#s3t-tablebucket-001), [S3T-TABLEBUCKET-053](../tablebucket-maintenance-metrics.md#s3t-tablebucket-053), [S3T-TABLEBUCKET-057](../tablebucket-maintenance-metrics.md#s3t-tablebucket-057), [S3T-TABLEBUCKET-093](../tablebucket-maintenance-metrics.md#s3t-tablebucket-093),
    [S3T-TABLEBUCKET-158](../tablebucket.md#s3t-tablebucket-158), [S3T-TABLEBUCKET-094](../tablebucket-maintenance-metrics.md#s3t-tablebucket-094), [S3T-TABLEBUCKET-095](../tablebucket-maintenance-metrics.md#s3t-tablebucket-095), [S3T-TABLEBUCKET-099](../tablebucket-maintenance-metrics.md#s3t-tablebucket-099), [S3T-TABLEBUCKET-113](../tablebucket-maintenance-metrics.md#s3t-tablebucket-113),
    [S3T-TABLEBUCKET-126](../tablebucket-encryption-storage.md#s3t-tablebucket-126), [S3T-NAMESPACE-030](../namespace.md#s3t-namespace-030), [S3T-TABLE-179](../table-maintenance.md#s3t-table-179), [S3T-TABLE-143](../table.md#s3t-table-143), [S3T-TABLE-267](../table.md#s3t-table-267), [S3T-TABLE-027](../table.md#s3t-table-027) ·
    hypotheses: H-S-091, H-S-096 · evidence: tablebucket/error-taxonomy/subresource-arn-variants

## Notes

H-S-091 qualified: upper-case name and trailing slash -> BadRequestException 400 'The specified bucket name
isn't valid...' (confirmed, no lower-casing), but an underscore name passes the ARN parser and is looked up:
NotFoundException 404 'The specified bucket does not exist.' (refuted for that cell). H-S-096 confirmed: table
ARN -> 400 'The specified bucket name isn't valid' (the '/table/<uuid>' suffix is treated as part of the name,
never resolved to the parent), empty region and service s3 -> 400 'The specified table bucket ARN is not
valid.'; extra cells: other region -> 400 'The provided Region does not match the Region of the resource.',
partition aws-cn -> 400 'The provided Partition does not match the Partition of the resource.', bare name ->
400 ARN not valid. Canonical Get afterwards -> 404 'No such metric configuration exists' (nothing created).
One call per variant.

Full original title: Metrics Put with ARN variants: upper-case name, trailing slash, table ARN, empty region,
service s3, other region, aws-cn, bare name -> 400; underscore name -> 404
