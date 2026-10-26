<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-116: DeleteTableBucket with 5 sub-resources set -> 204; same-name re-create gets a new tableBucketId, every sub-resource reads fresh (+0/+30 s)
_Full entry and notes of one finding; its summary entry is in
[tablebucket-replication.md](../tablebucket-replication.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-116"></a>**S3T-TABLEBUCKET-116** `identity` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **DeleteTableBucket with 5 sub-resources set -> 204; same-name re-create gets a new tableBucketId, every sub-resource reads fresh (+0/+30 s)**
  With encryption AES256, policy, maintenance enabled 7/3, INTELLIGENT_TIERING and metrics set (pre-delete
  reads {'encryption': "{'sseAlgorithm': 'AES256'}", 'policy': 'policy(235 chars)', 'metrics':
  'id=EntireBucket', 'replication': 'NotFoundException/404', 'maintenance': 'enabled/7/3', 'storage_class':
  'INTELLIGENT_TIERING'}), DeleteTableBucket -> HTTP 204; GetTableBucket -> NotFoundException after 0.0 s.
  CreateTableBucket with the same name (after the 409 window): the new ARN equals the old one, tableBucketId
  changed: True. Sub-resource Gets right after re-create: {'encryption': "{'sseAlgorithm': 'AES256'}",
  'policy': 'NotFoundException/404', 'metrics': 'NotFoundException/404', 'replication':
  'NotFoundException/404', 'maintenance': 'enabled/3/10', 'storage_class': 'STANDARD'}; at +30 s:
  {'encryption': "{'sseAlgorithm': 'AES256'}", 'policy': 'NotFoundException/404', 'metrics':
  'NotFoundException/404', 'replication': 'NotFoundException/404', 'maintenance': 'enabled/3/10',
  'storage_class': 'STANDARD'} (identical to +0).
  - ACK: is_arn_primary_key, custom_find · ops: DeleteTableBucket, CreateTableBucket, GetTableBucket,
    GetTableBucketEncryption, GetTableBucketPolicy, GetTableBucketMetricsConfiguration,
    GetTableBucketReplication, GetTableBucketMaintenanceConfiguration, GetTableBucketStorageClass · fields:
    tableBucketId, arn
  - repro: set 5 sub-resources -> DeleteTableBucket -> wait 404 -> CreateTableBucket(same name) until 201 ->
    GetTableBucket -> six Gets at +0 and +30 s
  - measurements: get_bucket_404_after_delete_s=0.0, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-097](../tablebucket-replication.md#s3t-tablebucket-097), [S3T-TABLEBUCKET-101](../tablebucket-replication.md#s3t-tablebucket-101), [S3T-TABLEBUCKET-092](../tablebucket-replication.md#s3t-tablebucket-092), [S3T-TABLEBUCKET-060](../tablebucket.md#s3t-tablebucket-060),
    [S3T-TABLEBUCKET-117](../tablebucket-encryption-storage.md#s3t-tablebucket-117) · hypotheses: H-S-022, H-S-099 · evidence:
    tablebucket/identity/subresources-across-recreate

## Notes

H-S-022 confirmed (no sub-resource blocked deletion; replication not set; same-name re-create needed 45
attempts because of the 409 window, see the delete-semantics finding). H-S-099 confirmed: no old
policy/metrics/maintenance/storage-class/encryption value survived on the reused ARN.

Full original title: DeleteTableBucket with 5 sub-resources set -> HTTP 204; same-name re-create gets a NEW
tableBucketId and reads every sub-resource as fresh (0 leaked) at +0 and +30 s
