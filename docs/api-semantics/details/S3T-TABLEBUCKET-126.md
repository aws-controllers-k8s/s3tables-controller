<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-126: Other-region bucket ARN on this endpoint -> 11 bucket ops (Get*/Put/Tag/Untag/ListNamespaces/Delete) all 400 region mismatch, both ways
_Full entry and notes of one finding; its summary entry is in
[tablebucket-encryption-storage.md](../tablebucket-encryption-storage.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-126"></a>**S3T-TABLEBUCKET-126** `cross-region` · impact high · unhandled (not handled in controller) · verified 2026-10-14
  **Other-region bucket ARN on this endpoint -> 11 bucket ops (Get*/Put/Tag/Untag/ListNamespaces/Delete) all 400 region mismatch, both ways**
  With bucket X existing in BOTH regions under the same name: the us-west-2 endpoint given the us-east-1 ARN
  -> {'BadRequestException (HTTP 400) The provided Region does not match the Region of the resource.':
  ['GetTableBucket', 'GetTableBucketPolicy', 'GetTableBucketEncryption',
  'GetTableBucketMaintenanceConfiguration', 'GetTableBucketStorageClass', 'PutTableBucketStorageClass',
  'ListNamespaces', 'ListTagsForResource', 'TagResource', 'UntagResource', 'DeleteTableBucket']}; the
  us-east-1 endpoint given the us-west-2 ARN -> {'BadRequestException (HTTP 400) The provided Region does not
  match the Region of the resource.': ['GetTableBucket', 'GetTableBucketPolicy', 'GetTableBucketEncryption',
  'GetTableBucketMaintenanceConfiguration', 'GetTableBucketStorageClass', 'PutTableBucketStorageClass',
  'ListNamespaces', 'ListTagsForResource', 'TagResource', 'UntagResource', 'DeleteTableBucket']}. Afterwards
  the us-east-1 bucket still exists (HTTP 200) with tags {'ackq-run': '5443b8', 'side': 'B'} (no foreign tag
  landed, 'side' tag not removed); the us-west-2 bucket HTTP 200, tags {'ackq-run': '5443b8', 'side': 'A'},
  storage class {'storageClassConfiguration': {'storageClass': 'STANDARD'}}. Cross-region DeleteTableBucket
  deleted nothing (B deleted: False, A deleted: False).
  - ACK: terminal_codes, is_arn_primary_key · ops: DeleteTableBucket, GetTableBucket,
    GetTableBucketEncryption, GetTableBucketMaintenanceConfiguration, GetTableBucketPolicy,
    GetTableBucketStorageClass, ListNamespaces, ListTagsForResource, PutTableBucketStorageClass, TagResource,
    UntagResource · fields: tableBucketARN, resourceArn
  - repro: CreateTableBucket(X) in both regions; from each endpoint call the 11 bucket ops with the OTHER
    region's ARN; verify both buckets via their own region
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-NAMESPACE-030](../namespace.md#s3t-namespace-030), [S3T-TABLE-179](../table-maintenance.md#s3t-table-179), [S3T-TABLEBUCKET-057](../tablebucket-maintenance-metrics.md#s3t-tablebucket-057), [S3T-TABLEBUCKET-100](../tablebucket-maintenance-metrics.md#s3t-tablebucket-100) · hypotheses: H-B-062,
    H-B-063, H-S-070 · evidence: tablebucket/cross-region/same-name-two-regions-arn-routing

## Notes

H-B-062 routing part confirmed, H-B-063 (contrarian: routes by ARN, cross-region delete works) refuted,
H-S-070 confirmed (BadRequestException, not 404, no redirect) for GetTableBucketPolicy,
PutTableBucketStorageClass, GetTableBucketEncryption/Maintenance/StorageClass, Tag/Untag/ListTags,
ListNamespaces, DeleteTableBucket; extends [S3T-NAMESPACE-030](../namespace.md#s3t-namespace-030). A region typo in a stored ARN is a terminal 400,
never 'resource missing'.

Full original title: Other-region bucket ARN on this endpoint: all 11 bucket ops
(Get*/Put/Tag/Untag/ListNamespaces/Delete) -> BadRequestException (HTTP 400) 'The provided Region does not
match the Region of the resource.' in both directions; nothing routed
