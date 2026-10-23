<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-060: GetTableBucket raw JSON keys are identical before and the same after setting 5 sub-resources (type=customer)
_Full entry and notes of one finding; its summary entry is in [tablebucket.md](../tablebucket.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-tablebucket-060"></a>**S3T-TABLEBUCKET-060** `read-gap` · impact medium · handled · verified 2026-10-13
  **GetTableBucket raw JSON keys are identical before and the same after setting 5 sub-resources (type=customer)**
  GetTableBucket on the fresh bucket returned raw JSON keys ['arn', 'createdAt', 'name', 'ownerAccountId',
  'tableBucketId', 'type'] (type=customer). After PutTableBucketEncryption(AES256), PutTableBucketPolicy,
  PutTableBucketMetricsConfiguration, PutTableBucketStorageClass(INTELLIGENT_TIERING) and
  PutTableBucketMaintenanceConfiguration(disabled 2/2) (outcomes {'PutTableBucketEncryption': 'HTTP 200',
  'PutTableBucketPolicy': 'HTTP 200', 'PutTableBucketMetricsConfiguration': 'HTTP 204',
  'PutTableBucketStorageClass': 'HTTP 200', 'PutTableBucketMaintenanceConfiguration': 'HTTP 204'})
  GetTableBucket returned keys ['arn', 'createdAt', 'name', 'ownerAccountId', 'tableBucketId', 'type']: none
  of the sub-resources is reflected. Replication was not set.
  - ACK: custom_find, hook:sdk_read_one_post_set_output · ops: GetTableBucket · fields: arn, name,
    ownerAccountId, createdAt, tableBucketId, type
  - repro: GetTableBucket (raw) -> set encryption/policy/metrics/storage class/maintenance -> GetTableBucket
    (raw) -> diff keys
  - handling: handled via `templates/hooks/table_bucket/sdk_read_one_post_set_output.go.tpl:6-12; pkg/resource/table_bucket/hooks.go:51-171`
  - related: [S3T-TABLEBUCKET-117](../tablebucket-encryption-storage.md#s3t-tablebucket-117), [S3T-TABLEBUCKET-116](../tablebucket-replication.md#s3t-tablebucket-116) · hypotheses: H-S-068, H-S-082 · evidence:
    tablebucket/sub-resources/maintenance-defaults-fill

## Notes

H-S-068 confirmed (5 of 6 sub-resources set; replication not tried). H-S-082 (static): the six Put ops are
modeled [create] with read-only read shapes; live here the maintenance Put merges into the stored value (see
the partial-Put finding) and the other Puts replace a single value, consistent with field-on-parent modeling
(fields on TableBucket, Put for create and update).
