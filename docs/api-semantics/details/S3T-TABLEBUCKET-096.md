<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-096: Model marks the 4 bucket sub-resource Deletes idempotent; on an absent config metrics/policy -> 204 but replication/encryption -> 404
_Full entry and notes of one finding; its summary entry is in
[tablebucket-replication.md](../tablebucket-replication.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-096"></a>**S3T-TABLEBUCKET-096** `codegen-artifact` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Model marks the 4 bucket sub-resource Deletes idempotent; on an absent config metrics/policy -> 204 but replication/encryption -> 404**
  botocore model 'idempotent': true on ['DeleteTableBucketMetricsConfiguration', 'DeleteTableBucketPolicy',
  'DeleteTableBucketEncryption', 'DeleteTableBucketReplication']. Live on a fresh bucket with nothing
  configured: DeleteTableBucketMetricsConfiguration -> HTTP 204, DeleteTableBucketPolicy -> HTTP 204,
  DeleteTableBucketReplication -> NotFoundException/404, DeleteTableBucketEncryption -> NotFoundException/404;
  a second DeleteTableBucketMetricsConfiguration right after a successful one -> HTTP 204.
  - ACK: exceptions.404, custom_delete · ops: DeleteTableBucketMetricsConfiguration, DeleteTableBucketPolicy,
    DeleteTableBucketEncryption, DeleteTableBucketReplication
  - repro: inspect the service model; Delete each sub-resource on a fresh bucket; Delete metrics twice
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-177](../tablebucket-maintenance-metrics.md#s3t-tablebucket-177), [S3T-TABLEBUCKET-093](../tablebucket-maintenance-metrics.md#s3t-tablebucket-093), [S3T-TABLEBUCKET-095](../tablebucket-maintenance-metrics.md#s3t-tablebucket-095), [S3T-TABLEBUCKET-092](../tablebucket-replication.md#s3t-tablebucket-092),
    [S3T-TABLEBUCKET-098](../tablebucket-replication.md#s3t-tablebucket-098), [S3T-SERVICE-014](../service.md#s3t-service-014) · hypotheses: H-S-102 · evidence:
    tablebucket/sub-resources/metrics-lifecycle

## Notes

H-S-102 confirmed (trait misleading): the trait does not predict the per-noun outcome; read the
Delete-when-absent code from this probe, not from the model.

Full original title: Model marks the four bucket sub-resource Deletes idempotent; observed on an absent
config: metrics HTTP 204, policy HTTP 204, replication NotFoundException/404, encryption NotFoundException/404
