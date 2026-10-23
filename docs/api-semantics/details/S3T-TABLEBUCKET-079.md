<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-079: Get/Put maintenance configuration on a just-deleted or never-existing bucket -> NotFoundException 404 'The specified bucket does not exist.'
_Full entry and notes of one finding; its summary entry is in
[tablebucket-maintenance-metrics.md](../tablebucket-maintenance-metrics.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-079"></a>**S3T-TABLEBUCKET-079** `error-code` · impact medium · handled · verified 2026-10-13
  **Get/Put maintenance configuration on a just-deleted or never-existing bucket -> NotFoundException 404 'The specified bucket does not exist.'**
  Immediately after DeleteTableBucket returned 204 (bucket emptied first):
  GetTableBucketMaintenanceConfiguration -> NotFoundException (HTTP 404) 'The specified bucket does not
  exist.'; PutTableBucketMaintenanceConfiguration {enabled, 3/10} -> NotFoundException (HTTP 404) 'The
  specified bucket does not exist.'. Same Get on an ARN whose bucket never existed -> NotFoundException (HTTP 404)
  'The specified bucket does not exist.'. On the live bucket this Get always returned HTTP 200 (fresh and
  under concurrent Puts).
  - ACK: exceptions.404 · ops: GetTableBucketMaintenanceConfiguration, PutTableBucketMaintenanceConfiguration
    · fields: tableBucketARN
  - repro: DeleteTable -> DeleteNamespace -> DeleteTableBucket -> Get/Put maintenance on the same ARN; Get on
    a never-existing ARN
  - handling: handled via `generator.yaml:243-249; pkg/resource/table_bucket/hooks.go:268-278`
  - related: [S3T-TABLEBUCKET-020](../tablebucket-encryption-storage.md#s3t-tablebucket-020), [S3T-TABLEBUCKET-158](../tablebucket.md#s3t-tablebucket-158) · hypotheses: H-S-107 · evidence:
    tablebucket/weird-inputs/maintenance-validation

## Notes

H-S-107 part 1 (404 on this Get is an unambiguous bucket-gone signal because a live bucket never 404s)
confirmed.

Full original title: GetTableBucketMaintenanceConfiguration on a just-deleted bucket -> NotFoundException
(HTTP 404); on a never-existing name -> NotFoundException (HTTP 404); Put on deleted -> NotFoundException
(HTTP 404)
