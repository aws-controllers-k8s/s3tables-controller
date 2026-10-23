<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-097: Fresh bucket Gets: encryption/maintenance/storage-class 200 with defaults, policy/metrics/replication 404; missing bucket -> 404 bucket msg
_Full entry and notes of one finding; its summary entry is in
[tablebucket-replication.md](../tablebucket-replication.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-097"></a>**S3T-TABLEBUCKET-097** `error-code` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Fresh bucket Gets: encryption/maintenance/storage-class 200 with defaults, policy/metrics/replication 404; missing bucket -> 404 bucket msg** (hypothesis refuted; behavior confirmed)
  Fresh bucket (nothing configured): encryption -> HTTP 200; policy -> NotFoundException (HTTP 404) 'The
  specified bucket policy does not exist.'; metrics -> NotFoundException (HTTP 404) 'No such metric
  configuration exists'; replication -> NotFoundException (HTTP 404) 'The specified replication configuration
  does not exist.'; maintenance -> HTTP 200; storage_class -> HTTP 200. Same six Gets on an ARN whose bucket
  never existed: encryption -> NotFoundException (HTTP 404) 'The specified bucket does not exist.'; policy ->
  NotFoundException (HTTP 404) 'The specified bucket does not exist.'; metrics -> NotFoundException (HTTP 404)
  'The specified bucket does not exist.'; replication -> NotFoundException (HTTP 404) 'The specified bucket
  does not exist.'; maintenance -> NotFoundException (HTTP 404) 'The specified bucket does not exist.';
  storage_class -> NotFoundException (HTTP 404) 'The specified bucket does not exist.'. The config-absent 404
  messages ({'policy': 'The specified bucket policy does not exist.', 'metrics': 'No such metric configuration
  exists', 'replication': 'The specified replication configuration does not exist.'}) differ from the
  bucket-missing message(s) ['The specified bucket does not exist.'], so the message text is the in-band
  discriminator.
  - ACK: exceptions.404, custom_find · ops: GetTableBucketEncryption, GetTableBucketPolicy,
    GetTableBucketMetricsConfiguration, GetTableBucketReplication, GetTableBucketMaintenanceConfiguration,
    GetTableBucketStorageClass · fields: tableBucketARN
  - repro: CreateTableBucket -> six Get* sub-resource reads -> the same six on
    arn:...:bucket/<never-existed-name>
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-174](../tablebucket-maintenance-metrics.md#s3t-tablebucket-174), [S3T-TABLEBUCKET-175](../tablebucket-maintenance-metrics.md#s3t-tablebucket-175), [S3T-TABLEBUCKET-176](../tablebucket-maintenance-metrics.md#s3t-tablebucket-176), [S3T-TABLEBUCKET-093](../tablebucket-maintenance-metrics.md#s3t-tablebucket-093),
    [S3T-TABLEBUCKET-092](../tablebucket-replication.md#s3t-tablebucket-092), [S3T-TABLEBUCKET-094](../tablebucket-maintenance-metrics.md#s3t-tablebucket-094), [S3T-TABLEBUCKET-173](../tablebucket-maintenance-metrics.md#s3t-tablebucket-173), [S3T-TABLEBUCKET-053](../tablebucket-maintenance-metrics.md#s3t-tablebucket-053),
    [S3T-TABLEBUCKETPOLICY-011](../tablebucket-policy.md#s3t-tablebucketpolicy-011), [S3T-TABLEBUCKETREPLICATION-023](../tablebucket-replication.md#s3t-tablebucketreplication-023), [S3T-TABLEBUCKET-101](../tablebucket-replication.md#s3t-tablebucket-101), [S3T-TABLEBUCKET-116](../tablebucket-replication.md#s3t-tablebucket-116) ·
    hypotheses: H-S-002, H-S-098 · evidence: tablebucket/error-taxonomy/subresource-arn-variants

## Notes

H-S-002 refuted: GetTableBucketEncryption on a never-configured bucket returns HTTP 200 {sseAlgorithm:
AES256}, not 404, so there is no message to match for that noun. H-S-098 confirmed:
GetTableBucketMetricsConfiguration is 404 for both cases and the messages differ ('No such metric
configuration exists' vs 'The specified bucket does not exist.'); policy ('The specified bucket policy does
not exist.') and replication ('The specified replication configuration does not exist.') follow the same
pattern; maintenance and storage class never 404 on a live bucket. One call per cell.

Full original title: Sub-resource Gets on a fresh bucket: encryption/maintenance/storage-class 200 with
defaults, policy/metrics/replication 404 with noun messages; missing bucket -> 404 'The specified bucket does
not exist.'
