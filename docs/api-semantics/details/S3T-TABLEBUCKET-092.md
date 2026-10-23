<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-092: Fresh bucket: Get metrics 404 'No such metric configuration exists'; Delete metrics/policy -> 204; Delete replication/encryption -> 404
_Full entry and notes of one finding; its summary entry is in
[tablebucket-replication.md](../tablebucket-replication.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-092"></a>**S3T-TABLEBUCKET-092** `error-code` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Fresh bucket: Get metrics 404 'No such metric configuration exists'; Delete metrics/policy -> 204; Delete replication/encryption -> 404** (hypothesis refuted; behavior confirmed)
  On a table bucket with no sub-resource ever set: GetTableBucketMetricsConfiguration -> NotFoundException
  (HTTP 404) 'No such metric configuration exists'. DeleteTableBucketMetricsConfiguration -> HTTP 204;
  DeleteTableBucketPolicy -> HTTP 204; DeleteTableBucketReplication -> NotFoundException (HTTP 404) 'The
  specified replication configuration does not exist.'; DeleteTableBucketEncryption -> NotFoundException (HTTP 404)
  'The specified encryption configuration does not exist.'. Get metrics again -> NotFoundException (HTTP 404)
  'No such metric configuration exists'.
  - ACK: exceptions.404, custom_delete · ops: GetTableBucketMetricsConfiguration,
    DeleteTableBucketMetricsConfiguration, DeleteTableBucketPolicy, DeleteTableBucketReplication,
    DeleteTableBucketEncryption · fields: tableBucketARN
  - repro: CreateTableBucket -> Get metrics -> Delete metrics / policy / replication / encryption -> Get
    metrics
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-174](../tablebucket-maintenance-metrics.md#s3t-tablebucket-174), [S3T-TABLEBUCKET-175](../tablebucket-maintenance-metrics.md#s3t-tablebucket-175), [S3T-TABLEBUCKET-176](../tablebucket-maintenance-metrics.md#s3t-tablebucket-176), [S3T-TABLEBUCKET-093](../tablebucket-maintenance-metrics.md#s3t-tablebucket-093),
    [S3T-TABLEBUCKET-094](../tablebucket-maintenance-metrics.md#s3t-tablebucket-094), [S3T-TABLEBUCKET-097](../tablebucket-replication.md#s3t-tablebucket-097), [S3T-TABLEBUCKET-177](../tablebucket-maintenance-metrics.md#s3t-tablebucket-177), [S3T-TABLEBUCKET-096](../tablebucket-replication.md#s3t-tablebucket-096), [S3T-TABLEBUCKET-095](../tablebucket-maintenance-metrics.md#s3t-tablebucket-095),
    [S3T-TABLEBUCKETPOLICY-011](../tablebucket-policy.md#s3t-tablebucketpolicy-011), [S3T-TABLEBUCKETREPLICATION-023](../tablebucket-replication.md#s3t-tablebucketreplication-023), [S3T-TABLEBUCKET-098](../tablebucket-replication.md#s3t-tablebucket-098), [S3T-SERVICE-014](../service.md#s3t-service-014),
    [S3T-TABLEBUCKET-101](../tablebucket-replication.md#s3t-tablebucket-101), [S3T-TABLEBUCKET-116](../tablebucket-replication.md#s3t-tablebucket-116) · hypotheses: H-S-043, H-S-088, H-S-098 · evidence:
    tablebucket/sub-resources/metrics-lifecycle

## Notes

H-S-043 qualified: the Get returns NotFoundException 404 with a metrics-specific message (confirmed) but
DeleteTableBucketMetricsConfiguration on an absent configuration returns HTTP 204, not 404 (refuted). H-S-088
qualified: DeleteTableBucketPolicy and DeleteTableBucketMetricsConfiguration are idempotent (204 when nothing
is configured, confirmed) but DeleteTableBucketReplication returns NotFoundException 404 'The specified
replication configuration does not exist.' (refuted for replication) and DeleteTableBucketEncryption, which
the hypothesis takes as the idempotent reference, also returns 404 'The specified encryption configuration
does not exist.'. H-S-098 cell (a): config-absent message 'No such metric configuration exists'. Each cell
once on one fresh bucket.

Full original title: Fresh bucket: Get metrics -> 404 'No such metric configuration exists'; Delete metrics
and Delete policy -> 204; Delete replication and encryption -> 404
